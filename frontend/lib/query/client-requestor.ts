import { apiClient } from "./api-client";
import type { Requestor, RequestorConfig } from "./requestor";

/**
 * Global state for client-side authentication.
 * These are set by the AuthProvider when it mounts.
 */
let accessToken: string | undefined;
let authLoaded = false;
let authLoadPromise: Promise<void> | null = null;
let authLoadResolve: (() => void) | null = null;

/**
 * User info from WorkOS (email, name) - passed as headers to backend.
 */
let userInfo: { email: string; name: string } | undefined;

/**
 * Set the access token for client-side requests.
 * Called by AuthProvider when the token changes.
 */
export function setAccessToken(token: string | undefined): void {
  accessToken = token;
}

/**
 * Set user info for client-side requests.
 * Called by AuthProvider when user data is available.
 */
export function setUserInfo(info: { email: string; name: string }): void {
  userInfo = info;
}

/**
 * Signal that auth has finished loading (whether authenticated or not).
 * Called by AuthProvider after initial auth check completes.
 */
export function setAuthLoaded(): void {
  authLoaded = true;
  if (authLoadResolve) {
    authLoadResolve();
    authLoadResolve = null;
    authLoadPromise = null;
  }
}

/**
 * Reset auth state. Used for sign-out.
 */
export function resetAuth(): void {
  accessToken = undefined;
  authLoaded = false;
  authLoadPromise = null;
  authLoadResolve = null;
}

/**
 * Get the access token, waiting for auth to load if necessary.
 * Returns undefined if user is not authenticated.
 */
export async function getToken(): Promise<string | undefined> {
  if (authLoaded) {
    return accessToken;
  }

  // Create a promise that resolves when auth loads
  if (!authLoadPromise) {
    authLoadPromise = new Promise<void>((resolve) => {
      authLoadResolve = resolve;
    });
  }

  await authLoadPromise;
  return accessToken;
}

/**
 * Get the access token synchronously.
 * Returns undefined if auth hasn't loaded or user is not authenticated.
 */
export function getTokenSync(): string | undefined {
  return accessToken;
}

/**
 * Check if auth has loaded.
 */
export function isAuthLoaded(): boolean {
  return authLoaded;
}

/**
 * Client-side requestor that injects Bearer token authentication.
 * Throws an error if called during SSR.
 */
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
