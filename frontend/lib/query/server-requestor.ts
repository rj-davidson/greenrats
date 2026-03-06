import { serverApiClient } from "@/lib/query/api-client";
import type { Requestor, RequestorConfig } from "@/lib/query/requestor";
import { withAuth } from "@workos-inc/authkit-nextjs";

async function getServerToken(): Promise<string> {
  const { accessToken } = await withAuth({ ensureSignedIn: true });
  return accessToken;
}

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
