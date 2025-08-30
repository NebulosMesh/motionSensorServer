export interface IApiResponse<T = any> {
  success: boolean;
  error?: string;
  data?: T;
}
