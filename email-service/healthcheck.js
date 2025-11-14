#!/usr/bin/env node

/* eslint-disable */
/**
 * Health check script for Email Service
 *
 * This script verifies:
 * 1. Main application process is running
 * 2. Can connect to RabbitMQ
 * 3. Can connect to PostgreSQL database
 */

const { execSync } = require('child_process');
const amqp = require('amqplib');
const { Client } = require('pg');

const RABBITMQ_URL =
  process.env.RABBITMQ_URL ||
  `amqp://${process.env.RABBITMQ_USERNAME || 'admin'}:${process.env.RABBITMQ_PASSWORD || 'password'}@${process.env.RABBITMQ_HOST || 'rabbitmq'}:${process.env.RABBITMQ_PORT || '5672'}`;

const DB_CONFIG = {
  host: process.env.DB_HOST || 'postgres',
  port: parseInt(process.env.DB_PORT || '5432', 10),
  user: process.env.DB_USERNAME || 'admin',
  password: process.env.DB_PASSWORD || 'password',
  database: process.env.DB_NAME || 'email_service',
  connectionTimeoutMillis: 2000,
};

async function checkRabbitMQ() {
  try {
    const connection = await amqp.connect(RABBITMQ_URL);
    await connection.close();
    return true;
  } catch (error) {
    console.error('RabbitMQ check failed:', error.message);
    return false;
  }
}

async function checkPostgreSQL() {
  try {
    const client = new Client(DB_CONFIG);
    await client.connect();
    await client.query('SELECT 1');
    await client.end();
    return true;
  } catch (error) {
    console.error('PostgreSQL check failed:', error.message);
    return false;
  }
}

function checkProcess() {
  try {
    // Check if the main application process (dist/main.js) is running
    const result = execSync('pgrep -f "node.*dist/main.js"', {
      encoding: 'utf8',
      stdio: 'pipe',
    });
    return result.trim().length > 0;
  } catch (error) {
    // pgrep returns non-zero if no process found
    return false;
  }
}

async function healthCheck() {
  const checks = {
    process: false,
    rabbitmq: false,
    postgresql: false,
  };

  // Check if main process is running
  checks.process = checkProcess();

  // Check RabbitMQ
  checks.rabbitmq = await checkRabbitMQ();

  // Check PostgreSQL
  checks.postgresql = await checkPostgreSQL();

  // All checks must pass
  const isHealthy = checks.process && checks.rabbitmq && checks.postgresql;

  if (!isHealthy) {
    console.error('Health check failed:', checks);
    process.exit(1);
  }

  console.log('Health check passed:', checks);
  process.exit(0);
}

// Run health check with timeout
const timeout = setTimeout(() => {
  console.error('Health check timeout');
  process.exit(1);
}, 5000);

healthCheck().finally(() => {
  clearTimeout(timeout);
});
