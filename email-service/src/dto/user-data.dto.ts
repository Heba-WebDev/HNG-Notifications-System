import { IsString, IsOptional, IsObject, IsUrl } from 'class-validator';

/**
 * UserData DTO
 * Structured data for template variables
 */
export class UserDataDto {
  @IsString()
  name!: string;

  @IsUrl({ require_protocol: true }, { message: 'link must be a valid HTTP URL' })
  link!: string;

  @IsOptional()
  @IsObject()
  meta?: Record<string, any>;
}

