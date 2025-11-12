from django.test import TestCase
from rest_framework.test import APIClient
from rest_framework import status
from .models import Template
import uuid


class TemplateModelTests(TestCase):
    """Test Template model."""

    def test_create_template(self):
        """Test creating a basic template."""
        template = Template.objects.create(
            code="welcome",
            name="Welcome Email",
            language="en",
            subject="Welcome!",
            content="Hello {{name}}, welcome to our service!"
        )
        self.assertEqual(template.code, "welcome")
        self.assertEqual(template.version, 1)
        self.assertTrue(template.is_active)

    def test_unique_together_code_version(self):
        """Test that code+version is unique."""
        Template.objects.create(
            code="test",
            name="Test 1",
            language="en",
            subject="Test",
            content="Test {{value}}",
            version=1
        )
        
        # Creating another with same code and version should fail
        with self.assertRaises(Exception):  # IntegrityError
            Template.objects.create(
                code="test",
                name="Test 2",
                language="en",
                subject="Test",
                content="Test {{value}}",
                version=1
            )


class TemplateAPIListCreateTests(TestCase):
    """Test Template List and Create APIs."""

    def setUp(self):
        self.client = APIClient()
        self.list_url = '/api/v1/templates/'

    def test_list_templates_empty(self):
        """Test listing templates when none exist."""
        response = self.client.get(self.list_url)
        self.assertEqual(response.status_code, status.HTTP_200_OK)
        data = response.json()
        self.assertTrue(data['success'])
        self.assertEqual(len(data['data']), 0)
        self.assertIn('meta', data)

    def test_create_template(self):
        """Test creating a template."""
        payload = {
            "code": "signup",
            "name": "Signup Confirmation",
            "language": "en",
            "subject": "Confirm Your Email",
            "content": "Click {{link}} to confirm."
        }
        response = self.client.post(self.list_url, payload, format='json')
        self.assertEqual(response.status_code, status.HTTP_201_CREATED)
        data = response.json()
        self.assertTrue(data['success'])
        self.assertEqual(data['data']['code'], "signup")
        self.assertEqual(data['data']['version'], 1)

    def test_create_template_invalid_content(self):
        """Test creating template with invalid Django template syntax."""
        # Use a template tag that requires a closing tag but doesn't have one
        payload = {
            "code": "bad",
            "name": "Bad Template",
            "language": "en",
            "subject": "Bad",
            "content": "{% for item in items %}{{ item }}"  # Missing {% endfor %}
        }
        response = self.client.post(self.list_url, payload, format='json')
        # Should fail validation
        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)

    def test_version_auto_increment(self):
        """Test that creating another template with same code increments version."""
        # Create v1
        resp1 = self.client.post(self.list_url, {
            "code": "test",
            "name": "Test v1",
            "language": "en",
            "subject": "Test",
            "content": "{{value}}"
        }, format='json')
        self.assertEqual(resp1.status_code, status.HTTP_201_CREATED)
        data1 = resp1.json()
        self.assertEqual(data1['data']['version'], 1)

        # Create v2
        resp2 = self.client.post(self.list_url, {
            "code": "test",
            "name": "Test v2",
            "language": "en",
            "subject": "Test",
            "content": "Updated {{value}}"
        }, format='json')
        if resp2.status_code != status.HTTP_201_CREATED:
            print("Response 2 error:", resp2.json())
        self.assertEqual(resp2.status_code, status.HTTP_201_CREATED)
        data2 = resp2.json()
        self.assertEqual(data2['data']['version'], 2)

        # Check that v1 is now inactive
        v1 = Template.objects.get(code="test", language="en", version=1)
        self.assertFalse(v1.is_active)

        # Check that v2 is active
        v2 = Template.objects.get(code="test", language="en", version=2)
        self.assertTrue(v2.is_active)

    def test_list_templates_with_pagination(self):
        """Test listing templates with pagination metadata."""
        # Create 5 templates
        for i in range(5):
            Template.objects.create(
                code=f"template{i}",
                name=f"Template {i}",
                language="en",
                subject="Test",
                content=f"Content {i}"
            )

        response = self.client.get(self.list_url)
        data = response.json()
        self.assertTrue(data['success'])
        self.assertEqual(len(data['data']), 5)
        # Pagination meta will be included if paginator is used
        if 'total' in data.get('meta', {}):
            self.assertEqual(data['meta']['total'], 5)
            self.assertIn('page', data['meta'])
            self.assertIn('total_pages', data['meta'])


class TemplateVersionListTests(TestCase):
    """Test Template Versions API."""

    def setUp(self):
        self.client = APIClient()
        # Create multiple versions
        Template.objects.create(code="doc", name="Doc v1", language="en", subject="Doc", content="v1", version=1, is_active=False)
        Template.objects.create(code="doc", name="Doc v2", language="en", subject="Doc", content="v2", version=2, is_active=False)
        Template.objects.create(code="doc", name="Doc v3", language="en", subject="Doc", content="v3", version=3, is_active=True)

    def test_list_versions(self):
        """Test listing versions of a template."""
        response = self.client.get('/api/v1/templates/doc/versions/')
        self.assertEqual(response.status_code, status.HTTP_200_OK)
        data = response.json()
        self.assertTrue(data['success'])
        self.assertEqual(len(data['data']), 3)
        # Should be ordered by -version
        self.assertEqual(data['data'][0]['version'], 3)

    def test_list_versions_not_found(self):
        """Test listing versions for non-existent code."""
        response = self.client.get('/api/v1/templates/nonexistent/versions/')
        self.assertEqual(response.status_code, status.HTTP_200_OK)
        data = response.json()
        # Should return empty list, not error
        self.assertEqual(len(data['data']), 0)


class RenderTemplateTests(TestCase):
    """Test Template Rendering."""

    def setUp(self):
        self.client = APIClient()
        self.template = Template.objects.create(
            code="welcome",
            name="Welcome",
            language="en",
            subject="Welcome {{name}}!",
            content="Hello {{name}}, welcome to {{company}}!",
            version=1,
            is_active=True
        )
        self.render_url = '/api/v1/templates/render/'

    def test_render_template_success(self):
        """Test rendering a template with variables."""
        payload = {
            "code": "welcome",
            "variables": {
                "name": "Alice",
                "company": "ACME Corp"
            }
        }
        response = self.client.post(self.render_url, payload, format='json')
        self.assertEqual(response.status_code, status.HTTP_200_OK)
        data = response.json()
        self.assertTrue(data['success'])
        self.assertIn("Alice", data['data']['content'])
        self.assertIn("ACME Corp", data['data']['content'])

    def test_render_template_missing_code(self):
        """Test rendering without code."""
        payload = {"variables": {}}
        response = self.client.post(self.render_url, payload, format='json')
        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)

    def test_render_template_not_found(self):
        """Test rendering non-existent template."""
        payload = {
            "code": "nonexistent",
            "variables": {}
        }
        response = self.client.post(self.render_url, payload, format='json')
        self.assertEqual(response.status_code, status.HTTP_404_NOT_FOUND)

    def test_render_template_with_language(self):
        """Test rendering template in specific language."""
        # Create Portuguese version
        Template.objects.create(
            code="welcome",
            name="Welcome PT",
            language="pt",
            subject="Bem-vindo {{name}}!",
            content="Olá {{name}}, bem-vindo!",
            version=1,
            is_active=True
        )

        payload = {
            "code": "welcome",
            "language": "pt",
            "variables": {"name": "Bruno"}
        }
        response = self.client.post(self.render_url, payload, format='json')
        self.assertEqual(response.status_code, status.HTTP_200_OK)
        data = response.json()
        self.assertIn("Olá", data['data']['content'])
        self.assertEqual(data['data']['language'], 'pt')

    def test_render_template_inactive(self):
        """Test that inactive templates cannot be rendered."""
        self.template.is_active = False
        self.template.save()

        payload = {
            "code": "welcome",
            "variables": {}
        }
        response = self.client.post(self.render_url, payload, format='json')
        self.assertEqual(response.status_code, status.HTTP_404_NOT_FOUND)


class HealthCheckTests(TestCase):
    """Test Health Check endpoint."""

    def setUp(self):
        self.client = APIClient()
        self.health_url = '/health/'

    def test_health_check_success(self):
        """Test health check when DB is healthy."""
        response = self.client.get(self.health_url)
        self.assertEqual(response.status_code, status.HTTP_200_OK)
        data = response.json()
        self.assertEqual(data['service'], 'template_service')
        self.assertEqual(data['status'], 'ok')
        self.assertEqual(data['database'], 'ok')

    def test_health_check_response_format(self):
        """Test health check response structure."""
        response = self.client.get(self.health_url)
        data = response.json()
        self.assertIn('service', data)
        self.assertIn('status', data)
        self.assertIn('database', data)
