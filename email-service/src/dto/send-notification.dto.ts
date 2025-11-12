import { IsEnum, IsNotEmpty, IsObject, IsString, IsOptional, IsInt, Min, ValidateNested } from 'class-validator';
import { Type } from 'class-transformer';
import { NotificationType } from './notification-type.enum';
import { UserDataDto } from './user-data.dto';

/**
 * SendNotificationDto
 */
export class SendNotificationDto {
  @IsEnum(NotificationType)
  notification_type!: NotificationType;

  @IsString()
  @IsNotEmpty()
  user_id!: string;

  @IsString()
  @IsNotEmpty()
  template_code!: string;

  @ValidateNested()
  @Type(() => UserDataDto)
  variables!: UserDataDto;

  @IsOptional()
  @IsString()
  request_id?: string;

  @IsOptional()
  @IsInt()
  @Min(0)
  priority?: number;

  @IsOptional()
  @IsObject()
  metadata?: Record<string, any>;
}

