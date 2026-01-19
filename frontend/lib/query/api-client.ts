import { PUBLIC_BACKEND_URL, PRIVATE_BACKEND_URL } from "@/lib/env";

/**
 * Configuration for API requests with token-based authentication.
 */
export interface RequestConfig {
  token?: string;
  baseUrl?: string;
  params?: Record<string, string>;
  /** User info to pass as headers (WorkOS access tokens don't include email/name) */
  userInfo?: { email: string; name: string };
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
  const { token, baseUrl = PUBLIC_BACKEND_URL, params, userInfo, ...init } = options;

  let url = `${baseUrl}${endpoint}`;
  if (params) {
    const searchParams = new URLSearchParams(params);
    url += `?${searchParams.toString()}`;
  }

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  if (init.headers) {
    const existingHeaders = new Headers(init.headers);
    existingHeaders.forEach((value, key) => {
      headers[key] = value;
    });
  }

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  if (userInfo?.email) {
    headers["X-User-Email"] = userInfo.email;
  }
  if (userInfo?.name) {
    headers["X-User-Name"] = userInfo.name;
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
