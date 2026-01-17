/**
 * Configuration options for requestor calls.
 */
export interface RequestorConfig {
  params?: Record<string, string>;
  signal?: AbortSignal;
}

/**
 * Requestor interface for dependency injection.
 * Both client-side and server-side requestors implement this interface,
 * allowing query options to be agnostic of the execution context.
 */
export type Requestor = {
  get: <T>(endpoint: string, config?: RequestorConfig) => Promise<T>;
  post: <T>(endpoint: string, body?: unknown, config?: RequestorConfig) => Promise<T>;
  put: <T>(endpoint: string, body?: unknown, config?: RequestorConfig) => Promise<T>;
  patch: <T>(endpoint: string, body?: unknown, config?: RequestorConfig) => Promise<T>;
  del: <T>(endpoint: string, config?: RequestorConfig) => Promise<T>;
};
