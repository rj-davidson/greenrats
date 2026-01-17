import { PUBLIC_BACKEND_URL, PRIVATE_BACKEND_URL } from "@/lib/env";

/**
 * Configuration for API requests with token-based authentication.
 */
export interface RequestConfig {
  token?: string;
  baseUrl?: string;
  params?: Record<string, string>;
}

/**
 * Error class for API request failures.
 * Includes status code, response data, and request details for debugging.
 */
export class APIError extends Error {
  constructor(
    public status: number,
    message: string,
    public response?: unknown,
    public url?: string,
    public method?: string,
  ) {
    super(message);
    this.name = "APIError";
  }
}

/**
 * Core API client with Bearer token authentication support.
 * Used by both client-side and server-side requestors.
 */
export async function apiClient<T>(
  endpoint: string,
  options: RequestInit & RequestConfig = {},
): Promise<T> {
  const { token, baseUrl = PUBLIC_BACKEND_URL, params, ...init } = options;

  let url = `${baseUrl}${endpoint}`;
  if (params) {
    const searchParams = new URLSearchParams(params);
    url += `?${searchParams.toString()}`;
  }

  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...init.headers,
  };

  if (token) {
    (headers as Record<string, string>)["Authorization"] = `Bearer ${token}`;
  }

  const response = await fetch(url, {
    ...init,
    headers,
  });

  if (!response.ok) {
    const data = await response.json().catch(() => null);
    throw new APIError(
      response.status,
      data?.error || data?.message || response.statusText,
      data,
      url,
      init.method || "GET",
    );
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json();
}

/**
 * Server-side API client that uses the private backend URL.
 * Useful for server components that can reach internal services.
 */
export async function serverApiClient<T>(
  endpoint: string,
  options: RequestInit & RequestConfig = {},
): Promise<T> {
  return apiClient<T>(endpoint, {
    ...options,
    baseUrl: options.baseUrl || PRIVATE_BACKEND_URL,
  });
}
