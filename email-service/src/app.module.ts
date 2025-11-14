import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { AppController } from './app.controller';
import { AppService } from './app.service';
import { EmailLog } from './entities/email-log.entity';

@Module({
  imports: [
    TypeOrmModule.forRoot({
      type: 'postgres',
      host: process.env.DB_HOST || 'localhost',
      port: parseInt(process.env.DB_PORT || '5432', 10),
      username: process.env.DB_USERNAME || 'admin',
      password: process.env.DB_PASSWORD || '',
      database: process.env.DB_NAME || 'email_service',
      entities: [EmailLog],
      synchronize: process.env.NODE_ENV !== 'production', // false in production
    }),
    TypeOrmModule.forFeature([EmailLog]),
  ],
  controllers: [AppController],
  providers: [AppService],
})
export class AppModule {}
