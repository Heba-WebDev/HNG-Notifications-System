function getRabbitMQUrl(): string {
  // Use full URL if provided
  if (process.env.RABBITMQ_URL) {
    return process.env.RABBITMQ_URL;
  }

  // Build URL from individual components
  const protocol = process.env.RABBITMQ_PROTOCOL || 'amqp';
  const username = process.env.RABBITMQ_USERNAME || 'admin';
  const password = process.env.RABBITMQ_PASSWORD || '';
  const host = process.env.RABBITMQ_HOST || 'localhost';
  const port = process.env.RABBITMQ_PORT || '5672';

  // In production, require password to be set
  if (process.env.NODE_ENV === 'production' && !password) {
    throw new Error('RABBITMQ_PASSWORD or RABBITMQ_URL must be set in production');
  }

  // In development, allow default password for local testing
  const finalPassword = password || (process.env.NODE_ENV === 'production' ? '' : 'password');

  return `${protocol}://${username}:${finalPassword}@${host}:${port}`;
}

export const RABBITMQ_CONFIG = {
  urls: [getRabbitMQUrl()],
  exchange: process.env.RABBITMQ_EXCHANGE || 'notifications.direct',
  queues: {
    email: process.env.RABBITMQ_QUEUE_EMAIL || 'email.queue',
    push: process.env.RABBITMQ_QUEUE_PUSH || 'push.queue',
    failed: process.env.RABBITMQ_QUEUE_FAILED || 'failed.queue',
    user: process.env.RABBITMQ_QUEUE_USER || 'user_queue',
    template: process.env.RABBITMQ_QUEUE_TEMPLATE || 'template_queue',
    auth: process.env.RABBITMQ_QUEUE_AUTH || 'auth_queue',
  },
  routingKeys: {
    email: 'email',
    push: 'push',
  },
};

