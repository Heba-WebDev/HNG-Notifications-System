# Template Service

A production-ready Django REST API microservice for managing notification templates with versioning, multi-language support, and variable substitution.

## Quick Start

### 1. Setup Environment
```bash
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt
```

### 2. Run Migrations
```bash
python manage.py makemigrations
python manage.py migrate
```

### 3. Run Tests
```bash
python manage.py test --verbosity=2
```

### 4. Start Server
```bash
python manage.py runserver
```

Visit:
- API: http://localhost:8000/api/v1/
- Docs: http://localhost:8000/api/v1/docs/
- Health: http://localhost:8000/health/

## Key Features

- ✅ **Template Versioning** - Automatic version increments with automatic deactivation of old versions
- ✅ **Multi-Language** - Store and render templates in multiple languages
- ✅ **Variable Substitution** - Django template engine with `{{ variable }}` syntax
- ✅ **Standard Response Envelope** - All responses wrapped with success/data/error/message/meta
- ✅ **Pagination** - Built-in pagination with metadata
- ✅ **Full CRUD** - Create, read, update, delete templates
- ✅ **Input Validation** - Template syntax validation on save
- ✅ **Health Checks** - Service health endpoint with DB connectivity check
- ✅ **API Documentation** - OpenAPI/Swagger and ReDoc
- ✅ **Production Ready** - Dockerfile, CI/CD, gunicorn WSGI server

## API Examples

### Create Template
```bash
curl -X POST http://localhost:8000/api/v1/templates/ \
  -H "Content-Type: application/json" \
  -d '{
    "code": "welcome",
    "name": "Welcome Email",
    "language": "en",
    "subject": "Welcome {{name}}!",
    "content": "Hello {{name}}, welcome to {{company}}!"
  }'
```

### Render Template
```bash
curl -X POST http://localhost:8000/api/v1/templates/render/ \
  -H "Content-Type: application/json" \
  -d '{
    "code": "welcome",
    "language": "en",
    "variables": {
      "name": "Alice",
      "company": "ACME Corp"
    }
  }'
```

### List Templates
```bash
curl http://localhost:8000/api/v1/templates/
```

### List Template Versions
```bash
curl http://localhost:8000/api/v1/templates/welcome/versions/
```

## Configuration

### Environment Variables
```bash
DEBUG=False  # Production should be False
DATABASE_URL=postgresql://user:pass@localhost/template_service
ALLOWED_HOSTS=localhost,127.0.0.1
SECRET_KEY=your-secret-key-here
```

### Settings
Edit `template_service/settings.py`:
- `DEBUG` - Enable debug mode
- `ALLOWED_HOSTS` - Allowed hostnames
- `DATABASES` - Database configuration
- `REST_FRAMEWORK` - DRF settings (pagination, etc.)

## Database

Default: SQLite (for dev). Production: PostgreSQL recommended.

### Switching to PostgreSQL
```python
# template_service/settings.py
DATABASES = {
    'default': {
        'ENGINE': 'django.db.backends.postgresql',
        'NAME': 'template_service',
        'USER': 'postgres',
        'PASSWORD': 'password',
        'HOST': 'localhost',
        'PORT': '5432',
    }
}
```

Then migrate:
```bash
python manage.py migrate
```

## Docker

### Build
```bash
docker build -t template-service:latest .
```

### Run
```bash
docker run -p 8000:8000 \
  -e DEBUG=False \
  -e ALLOWED_HOSTS=localhost \
  template-service:latest
```

## Testing

### Run All Tests
```bash
python manage.py test
```

### Run Specific Test Class
```bash
python manage.py test template_app.tests.RenderTemplateTests
```

### Run with Coverage
```bash
pip install coverage
coverage run --source='.' manage.py test
coverage report
coverage html  # Generate HTML report
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/templates/` | List all templates (paginated) |
| POST | `/api/v1/templates/` | Create new template |
| GET | `/api/v1/templates/{id}/` | Get template by ID |
| PATCH | `/api/v1/templates/{id}/` | Update template (partial) |
| PUT | `/api/v1/templates/{id}/` | Update template (full) |
| DELETE | `/api/v1/templates/{id}/` | Delete template |
| GET | `/api/v1/templates/{code}/versions/` | List versions of template |
| POST | `/api/v1/templates/render/` | Render template with variables |
| GET | `/health/` | Health check |
| GET | `/api/v1/docs/` | Swagger UI |
| GET | `/api/v1/redoc/` | ReDoc |
| GET | `/api/v1/schema/` | OpenAPI Schema |

## Request/Response Format

### Standard Response
```json
{
  "success": true,
  "data": {},
  "error": null,
  "message": "Operation successful",
  "meta": {
    "total": 10,
    "limit": 20,
    "page": 1,
    "total_pages": 1,
    "has_next": false,
    "has_previous": false
  }
}
```

### Template Model
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "code": "welcome",
  "name": "Welcome Email",
  "language": "en",
  "subject": "Welcome {{name}}!",
  "content": "Hello {{name}}, welcome to {{company}}!",
  "version": 1,
  "is_active": true,
  "created_at": "2025-11-12T10:00:00Z",
  "updated_at": "2025-11-12T10:00:00Z"
}
```

## Troubleshooting

### Port Already in Use
```bash
# Change port
python manage.py runserver 8001
```

### Database Connection Error
```bash
# Check database is running and credentials
python manage.py dbshell
```

### Migration Issues
```bash
# Reset migrations (dev only!)
python manage.py migrate template_app zero
python manage.py migrate
```

## Performance Notes

- Pagination default: 20 items per page
- Rendered templates are not cached (can add Redis caching)
- Database queries use select_related/prefetch_related where applicable
- Use gunicorn with multiple workers for production load

## Security

- ✅ Template syntax validated on save
- ✅ No exception details leaked in errors
- ✅ Read-only fields (id, version, is_active, timestamps)
- ✅ Database constraints enforce unique versions per code+language
- ✅ Health endpoint public but doesn't reveal stack traces

## Contributing

1. Write tests for new features
2. Run `python manage.py test` to verify
3. Follow Django and DRF conventions
4. Update API documentation if adding endpoints

## License

MIT
