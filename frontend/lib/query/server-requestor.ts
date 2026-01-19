import { serverApiClient } from "@/lib/query/api-client";
import type { Requestor, RequestorConfig } from "@/lib/query/requestor";
import { withAuth } from "@workos-inc/authkit-nextjs";

/**
 * Get the access token from WorkOS for server-side requests.
 * Uses ensureSignedIn to require authentication.
 */
async function getServerToken(): Promise<string> {
  const { accessToken } = await withAuth({ ensureSignedIn: true });
  return accessToken;
}

/**
 * Server-side requestor that uses WorkOS withAuth() for authentication.
 * Uses the private backend URL for internal service communication.
 */
export const makeServerRequest: Requestor = {
  async get<T>(endpoint: string, config?: RequestorConfig): Promise<T> {
    const token = await getServerToken();
    return serverApiClient<T>(endpoint, {
      method: "GET",
      token,
      params: config?.params,
      signal: config?.signal,
    });
  },

  async post<T>(endpoint: string, body?: unknown, config?: RequestorConfig): Promise<T> {
    const token = await getServerToken();
    return serverApiClient<T>(endpoint, {
      method: "POST",
      body: body ? JSON.stringify(body) : undefined,
      token,
      params: config?.params,
      signal: config?.signal,
    });
  },

  async put<T>(endpoint: string, body?: unknown, config?: RequestorConfig): Promise<T> {
    const token = await getServerToken();
    return serverApiClient<T>(endpoint, {
      method: "PUT",
      body: body ? JSON.stringify(body) : undefined,
      token,
      params: config?.params,
      signal: config?.signal,
    });
  },

  async patch<T>(endpoint: string, body?: unknown, config?: RequestorConfig): Promise<T> {
    const token = await getServerToken();
    return serverApiClient<T>(endpoint, {
      method: "PATCH",
      body: body ? JSON.stringify(body) : undefined,
      token,
      params: config?.params,
      signal: config?.signal,
    });
  },

  async del<T>(endpoint: string, config?: RequestorConfig): Promise<T> {
    const token = await getServerToken();
    return serverApiClient<T>(endpoint, {
      method: "DELETE",
      token,
      params: config?.params,
      signal: config?.signal,
    });
  },
};

/**
 * Create a server requestor that doesn't require authentication.
 * Useful for public endpoints.
 */
export const makePublicServerRequest: Requestor = {
  async get<T>(endpoint: string, config?: RequestorConfig): Promise<T> {
    return serverApiClient<T>(endpoint, {
      method: "GET",
      params: config?.params,
      signal: config?.signal,
    });
  },

  async post<T>(endpoint: string, body?: unknown, config?: RequestorConfig): Promise<T> {
    return serverApiClient<T>(endpoint, {
      method: "POST",
      body: body ? JSON.stringify(body) : undefined,
      params: config?.params,
      signal: config?.signal,
    });
  },

  async put<T>(endpoint: string, body?: unknown, config?: RequestorConfig): Promise<T> {
    return serverApiClient<T>(endpoint, {
      method: "PUT",
      body: body ? JSON.stringify(body) : undefined,
      params: config?.params,
      signal: config?.signal,
    });
  },

  async patch<T>(endpoint: string, body?: unknown, config?: RequestorConfig): Promise<T> {
    return serverApiClient<T>(endpoint, {
      method: "PATCH",
      body: body ? JSON.stringify(body) : undefined,
      params: config?.params,
      signal: config?.signal,
    });
  },

  async del<T>(endpoint: string, config?: RequestorConfig): Promise<T> {
    return serverApiClient<T>(endpoint, {
      method: "DELETE",
      params: config?.params,
      signal: config?.signal,
    });
  },
};
