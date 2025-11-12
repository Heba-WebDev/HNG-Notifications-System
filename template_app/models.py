from django.db import models
import uuid

class Template(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    code = models.CharField(max_length=100)  # not unique (multiple versions + languages)
    name = models.CharField(max_length=200)
    language = models.CharField(max_length=10, default='en')
    subject = models.CharField(max_length=255, blank=True, null=True)
    content = models.TextField()
    version = models.IntegerField(default=1)
    is_active = models.BooleanField(default=True)
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    class Meta:
        unique_together = ('code', 'language', 'version')  # Unique by code+language+version
        ordering = ['-version']

    def __str__(self):
        return f"{self.code} (lang={self.language}, v{self.version})"
