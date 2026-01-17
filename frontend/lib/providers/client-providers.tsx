"use client";

import {
  setAccessToken,
  setAuthLoaded,
} from "@/lib/query/client-requestor";
import { getQueryClient } from "@/lib/query/get-query-client";
import { useAccessToken } from "@workos-inc/authkit-nextjs/components";
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { useEffect } from "react";

import { Toaster } from "@/components/shadcn/sonner";
import { TooltipProvider } from "@/components/shadcn/tooltip";

/**
 * Captures the WorkOS access token and makes it available globally
 * for the client-side requestor.
 */
function AuthProvider({ children }: { children: React.ReactNode }) {
  const { accessToken, loading } = useAccessToken();

  useEffect(() => {
    setAccessToken(accessToken);
    if (!loading) {
      setAuthLoaded();
    }
  }, [accessToken, loading]);

  return <>{children}</>;
}

/**
 * Provides the TanStack Query client to the application.
 */
function QueryProvider({ children }: { children: React.ReactNode }) {
  const queryClient = getQueryClient();

  return (
    <QueryClientProvider client={queryClient}>
      {children}
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
}

/**
 * Combines all client-side providers into a single wrapper.
 * Order: AuthProvider -> QueryProvider -> TooltipProvider -> Toaster
 */
export function ClientProviders({ children }: { children: React.ReactNode }) {
  return (
    <AuthProvider>
      <QueryProvider>
        <TooltipProvider>
          {children}
          <Toaster />
        </TooltipProvider>
      </QueryProvider>
    </AuthProvider>
  );
}
