import { env } from "@/lib/env";

export interface RequestConfig {
  token?: string;
  baseUrl?: string;
  params?: Record<string, string>;
  timeoutMs?: number;
  /** User info to pass as headers (WorkOS access tokens don't include email/name) */
  userInfo?: { email: string; name: string };
}

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

const DEFAULT_SERVER_API_TIMEOUT_MS = 8000;

function withTimeoutSignal(
  signal: AbortSignal | null | undefined,
  timeoutMs?: number,
): {
  signal: AbortSignal | null | undefined;
  didTimeout: () => boolean;
  cleanup: () => void;
} {
  if (!timeoutMs || timeoutMs <= 0) {
    return {
      signal,
      didTimeout: () => false,
      cleanup: () => {},
    };
  }

  const controller = new AbortController();
  let timedOut = false;

  const onAbort = () => {
    controller.abort(signal?.reason);
  };

  if (signal) {
    if (signal.aborted) {
      onAbort();
    } else {
      signal.addEventListener("abort", onAbort, { once: true });
    }
  }

  const timeoutId = setTimeout(() => {
    timedOut = true;
    controller.abort();
  }, timeoutMs);

  return {
    signal: controller.signal,
    didTimeout: () => timedOut,
    cleanup: () => {
      clearTimeout(timeoutId);
      signal?.removeEventListener("abort", onAbort);
    },
  };
}

export async function apiClient<T>(
  endpoint: string,
  options: RequestInit & RequestConfig = {},
): Promise<T> {
  const { token, baseUrl = env.NEXT_PUBLIC_API_URL, params, timeoutMs, userInfo, ...init } = options;

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

  const { signal, didTimeout, cleanup } = withTimeoutSignal(init.signal, timeoutMs);

  let response: Response;
  try {
    response = await fetch(url, {
      ...init,
      headers,
      signal: signal ?? undefined,
    });
  } catch (error) {
    if (didTimeout()) {
      throw new APIError(
        504,
        `request timed out after ${timeoutMs}ms`,
        undefined,
        url,
        init.method || "GET",
      );
    }
    throw error;
  } finally {
    cleanup();
  }

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

export async function serverApiClient<T>(
  endpoint: string,
  options: RequestInit & RequestConfig = {},
): Promise<T> {
  return apiClient<T>(endpoint, {
    ...options,
    baseUrl: options.baseUrl || env.PRIVATE_API_URL || env.NEXT_PUBLIC_API_URL,
    timeoutMs: options.timeoutMs ?? DEFAULT_SERVER_API_TIMEOUT_MS,
  });
}
