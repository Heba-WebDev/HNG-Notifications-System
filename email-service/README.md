# Email Service

A NestJS microservice for handling email notifications in a distributed notification system.

## Overview

This service is part of a microservices architecture where different services are built with different technologies:
- **Email Service**: NestJS (this service)
- **Push Service**: Go
- **Other Services**: Python Django, etc.

All services share the same infrastructure instances: **RabbitMQ**, **PostgreSQL**, and **Redis**.

## Features

- ✅ Asynchronous email processing via RabbitMQ
- ✅ Email template support with variable substitution
- ✅ Retry mechanism with exponential backoff
- ✅ Email logging and status tracking
- ✅ Health check endpoints
- ✅ Dead-letter queue support
- ✅ Fully independent (no external shared package dependencies)

## Prerequisites

- Node.js 18+
- Docker and Docker Compose (for infrastructure)
- PostgreSQL database
- RabbitMQ instance
- Email credentials (Gmail, SendGrid, etc.)

## Installation

```bash
npm install
```

## Configuration

Create a `.env` file in the root directory:

```env
# Node Environment
NODE_ENV=production

# RabbitMQ Configuration
RABBITMQ_URL=amqp://admin:password@localhost:5672
# OR use individual components:
RABBITMQ_USERNAME=admin
RABBITMQ_PASSWORD=password
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_QUEUE_EMAIL=email.queue

# PostgreSQL Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=admin
DB_PASSWORD=password
DB_NAME=email_service

# Email Service Configuration (REQUIRED)
EMAIL_USER=your-email@gmail.com
EMAIL_PASS=your-app-password
EMAIL_FROM=noreply@example.com
EMAIL_SERVICE=gmail
```

## Running the Service

### Development

```bash
# Start infrastructure (RabbitMQ, PostgreSQL, Redis)
docker-compose up rabbitmq postgres redis

# Run the service in development mode
npm run start:dev
```

### Production

```bash
# Build the application
npm run build

# Start the service
npm run start:prod
```

### Docker

```bash
# Build and start all services (including infrastructure)
docker-compose up --build

# Or start only the email service (assuming infrastructure is running)
docker-compose up email-service
```

## Architecture

### Shared Infrastructure

All microservices connect to the same instances of:
- **RabbitMQ**: Message broker for async communication
- **PostgreSQL**: Database (each service uses its own database)
- **Redis**: Caching and rate limiting

See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed information about how services share infrastructure.

### Message Flow

1. Notification service publishes message to RabbitMQ exchange
2. Message is routed to `email.queue`
3. Email service consumes message
4. Email is sent via configured email provider
5. Status is logged in PostgreSQL
6. Failed messages are retried (up to 3 times)
7. After max retries, message goes to dead-letter queue

## API Endpoints

### RabbitMQ Message Patterns

#### `notification.email` (Event Pattern)
Consumes email notification messages from RabbitMQ.

**Message Format:**
```json
{
  "request_id": "uuid",
  "user": {
    "id": "user-id",
    "email": "user@example.com"
  },
  "template": {
    "subject": "Welcome {{name}}",
    "body": "<h1>Hello {{name}}</h1><p>Click here: {{link}}</p>"
  },
  "variables": {
    "name": "John Doe",
    "link": "https://example.com/verify"
  }
}
```

#### `health.check` (Message Pattern)
Health check endpoint.

**Response:**
```json
{
  "success": true,
  "message": "Email service is healthy",
  "data": {
    "timestamp": "2024-01-01T00:00:00.000Z"
  }
}
```

#### `email.update_status` (Message Pattern)
Update email notification status.

**Request:**
```json
{
  "request_id": "uuid",
  "status": "sent|failed|bounced",
  "timestamp": "2024-01-01T00:00:00.000Z",
  "error": "Error message (optional)"
}
```

#### `email.get_by_user_id` (Message Pattern)
Get all email notifications for a user.

**Request:**
```json
{
  "user_id": "user-id"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Email notifications retrieved successfully",
  "data": [
    {
      "request_id": "uuid",
      "status": "sent",
      "subject": "Welcome",
      "error_message": null,
      "created_at": "2024-01-01T00:00:00.000Z"
    }
  ]
}
```

## Database Schema

### email_logs Table

```sql
CREATE TABLE email_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  request_id VARCHAR(255) UNIQUE NOT NULL,
  user_id VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL,
  subject VARCHAR(500) NOT NULL,
  body TEXT NOT NULL,
  status VARCHAR(50) DEFAULT 'pending',
  error_message TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Testing

```bash
# Unit tests
npm run test

# E2E tests
npm run test:e2e

# Test coverage
npm run test:cov
```

## Project Structure

```
email-service/
├── src/
│   ├── config/           # Configuration files
│   │   └── rabbitmq.config.ts
│   ├── dto/              # Data Transfer Objects
│   │   ├── notification-type.enum.ts
│   │   ├── response.dto.ts
│   │   ├── send-notification.dto.ts
│   │   └── user-data.dto.ts
│   ├── entities/         # TypeORM entities
│   │   └── email-log.entity.ts
│   ├── interfaces/       # TypeScript interfaces
│   │   └── pagination-meta.interface.ts
│   ├── app.controller.ts # Message handlers
│   ├── app.service.ts    # Business logic
│   ├── app.module.ts     # NestJS module
│   └── main.ts           # Application entry point
├── test/                 # Test files
├── docker-compose.yml    # Docker Compose configuration
├── Dockerfile           # Docker image definition
└── package.json        # Dependencies and scripts
```

## Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `NODE_ENV` | Environment mode | No | `development` |
| `RABBITMQ_URL` | Full RabbitMQ connection URL | No* | - |
| `RABBITMQ_HOST` | RabbitMQ host | No* | `localhost` |
| `RABBITMQ_PORT` | RabbitMQ port | No* | `5672` |
| `RABBITMQ_USERNAME` | RabbitMQ username | No* | `admin` |
| `RABBITMQ_PASSWORD` | RabbitMQ password | No* | `password` |
| `RABBITMQ_QUEUE_EMAIL` | Email queue name | No | `email.queue` |
| `DB_HOST` | PostgreSQL host | No | `localhost` |
| `DB_PORT` | PostgreSQL port | No | `5432` |
| `DB_USERNAME` | PostgreSQL username | No | `admin` |
| `DB_PASSWORD` | PostgreSQL password | No | - |
| `DB_NAME` | Database name | No | `email_service` |
| `EMAIL_USER` | Email account username | **Yes** | - |
| `EMAIL_PASS` | Email account password/app password | **Yes** | - |
| `EMAIL_FROM` | From email address | No | `EMAIL_USER` |
| `EMAIL_SERVICE` | Email service provider | No | `gmail` |

*Either `RABBITMQ_URL` or individual components (`RABBITMQ_HOST`, etc.) must be provided.

## Troubleshooting

### Email not sending
- Check `EMAIL_USER` and `EMAIL_PASS` are set correctly
- For Gmail, use an App Password, not your regular password
- Check email service provider settings

### Cannot connect to RabbitMQ
- Ensure RabbitMQ is running: `docker-compose up rabbitmq`
- Check connection string in environment variables
- Verify network connectivity if using Docker

### Database connection errors
- Ensure PostgreSQL is running: `docker-compose up postgres`
- Verify database credentials
- Check if database exists: `CREATE DATABASE email_service;`

## License

MIT
