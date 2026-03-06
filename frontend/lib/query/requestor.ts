export interface RequestorConfig {
  params?: Record<string, string>;
  signal?: AbortSignal;
}

export type Requestor = {
  get: <T>(endpoint: string, config?: RequestorConfig) => Promise<T>;
  post: <T>(endpoint: string, body?: unknown, config?: RequestorConfig) => Promise<T>;
  put: <T>(endpoint: string, body?: unknown, config?: RequestorConfig) => Promise<T>;
  patch: <T>(endpoint: string, body?: unknown, config?: RequestorConfig) => Promise<T>;
  del: <T>(endpoint: string, config?: RequestorConfig) => Promise<T>;
};
