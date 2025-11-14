import { PaginationMeta } from '../interfaces/pagination-meta.interface';

export interface ResponseDto<T> {
  success: boolean;
  data?: T;
  error?: string;
  message: string;
  meta?: PaginationMeta;
}

