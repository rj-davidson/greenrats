import { apiClient } from "@/lib/query/api-client";
import type { Requestor, RequestorConfig } from "@/lib/query/requestor";

let accessToken: string | undefined;
let authLoaded = false;
let authLoadPromise: Promise<void> | null = null;
let authLoadResolve: (() => void) | null = null;

let userInfo: { email: string; name: string } | undefined;

export function setAccessToken(token: string | undefined): void {
  accessToken = token;
}

export function setUserInfo(info: { email: string; name: string }): void {
  userInfo = info;
}

export function setAuthLoaded(): void {
  authLoaded = true;
  if (authLoadResolve) {
    authLoadResolve();
    authLoadResolve = null;
    authLoadPromise = null;
  }
}

export function resetAuth(): void {
  accessToken = undefined;
  authLoaded = false;
  authLoadPromise = null;
  authLoadResolve = null;
}

export async function getToken(): Promise<string | undefined> {
  if (authLoaded) {
    return accessToken;
  }

  if (!authLoadPromise) {
    authLoadPromise = new Promise<void>((resolve) => {
      authLoadResolve = resolve;
    });
  }

  await authLoadPromise;
  return accessToken;
}

export function getTokenSync(): string | undefined {
  return accessToken;
}

export function isAuthLoaded(): boolean {
  return authLoaded;
}

export const makeClientRequest: Requestor = {
  async get<T>(endpoint: string, config?: RequestorConfig): Promise<T> {
    if (typeof window === "undefined") {
      throw new Error("makeClientRequest cannot be used during SSR");
    }
    const token = await getToken();
    return apiClient<T>(endpoint, {
      method: "GET",
      token,
      userInfo,
      params: config?.params,
      signal: config?.signal,
    });
  },

  async post<T>(endpoint: string, body?: unknown, config?: RequestorConfig): Promise<T> {
    if (typeof window === "undefined") {
      throw new Error("makeClientRequest cannot be used during SSR");
    }
    const token = await getToken();
    return apiClient<T>(endpoint, {
      method: "POST",
      body: body ? JSON.stringify(body) : undefined,
      token,
      userInfo,
      params: config?.params,
      signal: config?.signal,
    });
  },

  async put<T>(endpoint: string, body?: unknown, config?: RequestorConfig): Promise<T> {
    if (typeof window === "undefined") {
      throw new Error("makeClientRequest cannot be used during SSR");
    }
    const token = await getToken();
    return apiClient<T>(endpoint, {
      method: "PUT",
      body: body ? JSON.stringify(body) : undefined,
      token,
      userInfo,
      params: config?.params,
      signal: config?.signal,
    });
  },

  async patch<T>(endpoint: string, body?: unknown, config?: RequestorConfig): Promise<T> {
    if (typeof window === "undefined") {
      throw new Error("makeClientRequest cannot be used during SSR");
    }
    const token = await getToken();
    return apiClient<T>(endpoint, {
      method: "PATCH",
      body: body ? JSON.stringify(body) : undefined,
      token,
      userInfo,
      params: config?.params,
      signal: config?.signal,
    });
  },

  async del<T>(endpoint: string, config?: RequestorConfig): Promise<T> {
    if (typeof window === "undefined") {
      throw new Error("makeClientRequest cannot be used during SSR");
    }
    const token = await getToken();
    return apiClient<T>(endpoint, {
      method: "DELETE",
      token,
      userInfo,
      params: config?.params,
      signal: config?.signal,
    });
  },
};
