from rest_framework import generics, status
from rest_framework.response import Response
from .models import Template
from .serializers import TemplateSerializer
from django.template import Template as DjangoTemplate, Context, TemplateSyntaxError
from rest_framework.views import APIView
from django.db import connection
from rest_framework.pagination import PageNumberPagination
import logging

logger = logging.getLogger(__name__)


def _envelope_response(success: bool, data=None, message: str = "", error: str = None, meta: dict = None):
    """Response envelope matching required spec."""
    payload = {
        "success": success,
        "data": data,
        "error": error,
        "message": message,
        "meta": meta or {}
    }
    return payload


def _build_pagination_meta(page_obj):
    """Build pagination meta from page object."""
    if page_obj is None:
        return {}
    
    try:
        paginator = page_obj.paginator
        current_page = page_obj.number
        
        return {
            "total": paginator.count,
            "limit": paginator.per_page,
            "page": current_page,
            "total_pages": paginator.num_pages,
            "has_next": page_obj.has_next(),
            "has_previous": page_obj.has_previous()
        }
    except AttributeError:
        return {}


class TemplateListCreateView(generics.ListCreateAPIView):
    queryset = Template.objects.all()
    serializer_class = TemplateSerializer
    pagination_class = PageNumberPagination

    def list(self, request, *args, **kwargs):
        queryset = self.filter_queryset(self.get_queryset())
        page = self.paginate_queryset(queryset)
        
        if page is not None:
            serializer = self.get_serializer(page, many=True)
            meta = _build_pagination_meta(page)
            return Response(_envelope_response(
                True, 
                data=serializer.data, 
                message="Templates retrieved", 
                meta=meta
            ))
        
        serializer = self.get_serializer(queryset, many=True)
        return Response(_envelope_response(
            True, 
            data=serializer.data, 
            message="Templates retrieved"
        ))

    def create(self, request, *args, **kwargs):
        """Override create to return enveloped response."""
        serializer = self.get_serializer(data=request.data)
        serializer.is_valid(raise_exception=True)
        self.perform_create(serializer)
        headers = self.get_success_headers(serializer.data)
        return Response(
            _envelope_response(True, data=serializer.data, message="Template created"),
            status=status.HTTP_201_CREATED,
            headers=headers
        )

    def perform_create(self, serializer):
        """Auto-increment version and deactivate older versions."""
        code = serializer.validated_data.get("code")
        language = serializer.validated_data.get("language", "en")

        latest = Template.objects.filter(code=code, language=language).order_by('-version').first()
        next_version = latest.version + 1 if latest else 1

        serializer.save(version=next_version)
        Template.objects.filter(code=code, language=language).exclude(
            version=next_version
        ).update(is_active=False)


class TemplateDetailView(generics.RetrieveUpdateDestroyAPIView):
    """Detail view for individual templates."""
    queryset = Template.objects.all()
    serializer_class = TemplateSerializer

    def retrieve(self, request, *args, **kwargs):
        """Override retrieve to return enveloped response."""
        instance = self.get_object()
        serializer = self.get_serializer(instance)
        return Response(
            _envelope_response(True, data=serializer.data, message="Template retrieved"),
            status=status.HTTP_200_OK
        )

    def update(self, request, *args, **kwargs):
        """Override update (PATCH/PUT) to return enveloped response."""
        partial = kwargs.pop('partial', False)
        instance = self.get_object()
        serializer = self.get_serializer(instance, data=request.data, partial=partial)
        serializer.is_valid(raise_exception=True)
        self.perform_update(serializer)
        return Response(
            _envelope_response(True, data=serializer.data, message="Template updated"),
            status=status.HTTP_200_OK
        )

    def destroy(self, request, *args, **kwargs):
        """Override destroy to return enveloped response."""
        instance = self.get_object()
        self.perform_destroy(instance)
        return Response(
            _envelope_response(True, data=None, message="Template deleted"),
            status=status.HTTP_204_NO_CONTENT
        )


class TemplateVersionListView(generics.ListAPIView):
    serializer_class = TemplateSerializer
    pagination_class = PageNumberPagination

    def get_queryset(self):
        code = self.kwargs['code']
        language = self.request.query_params.get('language', 'en')
        return Template.objects.filter(code=code, language=language).order_by('-version')

    def list(self, request, *args, **kwargs):
        """Override list to return enveloped response with pagination meta."""
        queryset = self.filter_queryset(self.get_queryset())
        page = self.paginate_queryset(queryset)
        
        if page is not None:
            serializer = self.get_serializer(page, many=True)
            meta = _build_pagination_meta(page)
            return Response(_envelope_response(
                True, 
                data=serializer.data, 
                message="Versions retrieved", 
                meta=meta
            ))
        
        serializer = self.get_serializer(queryset, many=True)
        return Response(_envelope_response(
            True, 
            data=serializer.data, 
            message="Versions retrieved"
        ))


class RenderTemplateView(APIView):
    def post(self, request, *args, **kwargs):
        code = request.data.get("code")
        variables = request.data.get("variables", {})
        language = request.data.get("language", "en") 

        if not code:
            return Response(
                _envelope_response(False, data=None, message="Missing code", error="missing_code"),
                status=status.HTTP_400_BAD_REQUEST
            )

        try:
            template = Template.objects.filter(
                code=code, 
                language=language, 
                is_active=True
            ).latest("version")
        except Template.DoesNotExist:
            return Response(
                _envelope_response(False, data=None, message="Template not found", error="not_found"),
                status=status.HTTP_404_NOT_FOUND
            )

        try:
            django_template = DjangoTemplate(template.content)
            rendered_content = django_template.render(Context(variables))
        except TemplateSyntaxError as e:
            logger.error(f"Template syntax error for code={code}: {str(e)}")
            return Response(
                _envelope_response(False, data=None, message="Template render error", error="render_error"),
                status=status.HTTP_500_INTERNAL_SERVER_ERROR
            )

        data = {
            "code": template.code,
            "subject": template.subject,
            "content": rendered_content,
            "version": template.version,
            "language": template.language
        }

        return Response(
            _envelope_response(True, data=data, message="Template rendered successfully"),
            status=status.HTTP_200_OK
        )


class HealthCheckView(APIView):
    authentication_classes = []
    permission_classes = []

    def get(self, request, *args, **kwargs):
        try:
            with connection.cursor() as cursor:
                cursor.execute("SELECT 1;")
                cursor.fetchone()
            db_status = "ok"
        except Exception as e:
            logger.error(f"Health check DB error: {str(e)}")
            db_status = "error"

        is_healthy = db_status == "ok"
        health_status = {
            "service": "template_service",
            "status": "ok" if is_healthy else "degraded",
            "database": db_status,
        }

        http_status = status.HTTP_200_OK if is_healthy else status.HTTP_503_SERVICE_UNAVAILABLE
        return Response(health_status, status=http_status)