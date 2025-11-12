import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { EmailLog } from './entities/email-log.entity';
import * as nodemailer from 'nodemailer';
import * as dotenv from 'dotenv';

// Load environment variables from .env file
dotenv.config();

@Injectable()
export class AppService {
  private transporter: nodemailer.Transporter;

  constructor(
    @InjectRepository(EmailLog)
    private readonly emailLogRepository: Repository<EmailLog>,
  ) {
    // Validate required email credentials
    const emailUser = process.env.EMAIL_USER;
    const emailPass = process.env.EMAIL_PASS;

    if (!emailUser || !emailPass) {
      if (process.env.NODE_ENV === 'production') {
        throw new Error('EMAIL_USER and EMAIL_PASS must be set in production');
      }
      console.warn('⚠️  EMAIL_USER and EMAIL_PASS not set. Email sending will fail.');
    }

    this.transporter = nodemailer.createTransport({
      service: process.env.EMAIL_SERVICE || 'gmail',
      auth: {
        user: emailUser || 'your-email@gmail.com',
        pass: emailPass || 'your-app-password',
      },
    });
  }

  async processEmailNotification(
    data: {
      request_id: string;
      user: any;
      template: any;
      variables: Record<string, any>;
    },
    retryCount: number = 0,
  ): Promise<void> {
    const maxRetries = 3;

    try {
      const existing = await this.emailLogRepository.findOne({
        where: { request_id: data.request_id },
      });

      if (existing) {
        console.log(`Email ${data.request_id} already processed`);
        return;
      }

      const emailLog = this.emailLogRepository.create({
        request_id: data.request_id,
        user_id: data.user.id,
        email: data.user.email,
        subject: data.template.subject,
        body: data.template.body,
        status: 'pending',
      });

      await this.emailLogRepository.save(emailLog);

      // Substitute variables
      let subject = data.template.subject;
      let body = data.template.body;

      Object.keys(data.variables).forEach((key) => {
        const regex = new RegExp(`{{${key}}}`, 'g');
        subject = subject.replace(regex, data.variables[key] || '');
        body = body.replace(regex, data.variables[key] || '');
      });

      const emailFrom = process.env.EMAIL_FROM || process.env.EMAIL_USER || 'noreply@example.com';
      
      await this.transporter.sendMail({
        from: emailFrom,
        to: data.user.email,
        subject: subject,
        html: body,
      });

      emailLog.status = 'sent';
      await this.emailLogRepository.save(emailLog);

      console.log(`Email sent successfully: ${data.request_id}`);
    } catch (error) {
      console.error(`Failed to send email: ${error.message}`);

      if (retryCount < maxRetries) {
        // Exponential backoff: 2^retryCount * 1000ms
        // Attempt 1: 1s, Attempt 2: 2s, Attempt 3: 4s
        const delay = Math.pow(2, retryCount) * 1000;
        console.log(
          `Retrying in ${delay}ms (attempt ${retryCount + 1}/${maxRetries + 1})`,
        );

        await new Promise((resolve) => setTimeout(resolve, delay));
        return this.processEmailNotification(data, retryCount + 1);
      }

      // Max retries reached
      const emailLog = await this.emailLogRepository.findOne({
        where: { request_id: data.request_id },
      });

      if (emailLog) {
        emailLog.status = 'failed';
        emailLog.error_message = error.message;
        await this.emailLogRepository.save(emailLog);
      }

      // Move to dead-letter queue after max retries
      await this.moveToDeadLetterQueue(data, error);

      throw error; // Re-throw to trigger RabbitMQ retry/dead-letter queue
    }
  }

  async checkDatabase(): Promise<void> {
    await this.emailLogRepository.query('SELECT 1');
  }

  async updateStatus(
    requestId: string,
    status: string,
    timestamp?: string,
    error?: string,
  ): Promise<EmailLog | null> {
    const emailLog = await this.emailLogRepository.findOne({
      where: { request_id: requestId },
    });

    if (!emailLog) {
      return null;
    }

    // Map status values: 'delivered' -> 'sent', keep others as is
    const mappedStatus = status === 'delivered' ? 'sent' : status;
    
    emailLog.status = mappedStatus;
    if (error) {
      emailLog.error_message = error;
    }

    await this.emailLogRepository.save(emailLog);

    return emailLog;
  }

  async getNotificationsByUserId(userId: string): Promise<any[]> {
    const notifications = await this.emailLogRepository.find({
      where: { user_id: userId },
      order: { created_at: 'DESC' },
    });

    return notifications.map((log) => ({
      request_id: log.request_id,
      status: log.status,
      subject: log.subject,
      error_message: log.error_message,
      created_at: log.created_at,
    }));
  }

  private async moveToDeadLetterQueue(data: any, error: Error): Promise<void> {
    console.error(
      `Moving to dead-letter queue: ${data.request_id}`,
      error.message,
    );
    // TODO: Implement dead-letter queue emission
  }
}
