import { Controller } from '@nestjs/common';
import { EventPattern, MessagePattern, Payload } from '@nestjs/microservices';
import { AppService } from './app.service';

@Controller()
export class AppController {
  constructor(private readonly appService: AppService) {}

  @EventPattern('notification.email')
  async handleEmailNotification(@Payload() data: any) {
    console.log('Received email notification:', data.request_id);

    try {
      await this.appService.processEmailNotification(data);
    } catch (error) {
      console.error('Failed to process email notification:', error);
      // Message will go to dead letter queue after retries
      throw error;
    }
  }

  @MessagePattern('health.check')
  async healthCheck() {
    try {
      await this.appService.checkDatabase();
      return {
        success: true,
        message: 'Email service is healthy',
        data: { timestamp: new Date().toISOString() },
      };
    } catch (error) {
      return {
        success: false,
        message: 'Email service is unhealthy',
        error: error.message || 'Internal server error',
      };
    }
  }

  @MessagePattern('email.update_status')
  async updateStatus(
    @Payload()
    data: {
      request_id: string;
      status: string;
      timestamp?: string;
      error?: string;
    },
  ) {
    try {
      const result = await this.appService.updateStatus(
        data.request_id,
        data.status,
        data.timestamp,
        data.error,
      );

      if (!result) {
        return {
          success: false,
          message: 'Email notification not found',
          error: 'NOT_FOUND',
        };
      }

      return {
        success: true,
        message: 'Email notification status updated successfully',
        data: result,
      };
    } catch (error) {
      return {
        success: false,
        message: 'Failed to update email notification status',
        error: error.message || 'Internal server error',
      };
    }
  }

  @MessagePattern('email.get_by_user_id')
  async getByUserId(@Payload() data: { user_id: string }) {
    try {
      const notifications = await this.appService.getNotificationsByUserId(
        data.user_id,
      );
      return {
        success: true,
        message: 'Email notifications retrieved successfully',
        data: notifications,
      };
    } catch (error) {
      return {
        success: false,
        message: 'Failed to retrieve email notifications',
        error: error.message || 'Internal server error',
      };
    }
  }
}
