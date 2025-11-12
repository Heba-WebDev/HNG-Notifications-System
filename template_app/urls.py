from django.urls import path
from . import views

urlpatterns = [
    path('api/v1/templates/', views.TemplateListCreateView.as_view(), name='template-list'),
    path('api/v1/templates/<uuid:pk>/', views.TemplateDetailView.as_view(), name='template-detail'),
    path('api/v1/templates/<str:code>/versions/', views.TemplateVersionListView.as_view(), name='template-versions'),
    path('api/v1/templates/render/', views.RenderTemplateView.as_view(), name='render-template'),
    path('health/', views.HealthCheckView.as_view(), name='health-check'),
]
