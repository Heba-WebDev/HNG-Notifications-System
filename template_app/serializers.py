from rest_framework import serializers
from .models import Template
from django.template import Template as DjangoTemplate, TemplateSyntaxError


class TemplateSerializer(serializers.ModelSerializer):
    class Meta:
        model = Template
        # Expose all model fields but make internal fields read-only so clients
        # cannot set them on create/update (versioning is managed server-side)
        fields = '__all__'
        read_only_fields = ('id', 'version', 'is_active', 'created_at', 'updated_at')
        # Remove unique_together from validators since we'll handle versioning in perform_create
        validators = []

    def validate_content(self, value):
        """Validate that the template content is valid Django template syntax."""
        try:
            DjangoTemplate(value)
        except TemplateSyntaxError as e:
            raise serializers.ValidationError(
                f"Invalid template syntax: {str(e)}"
            )
        return value
