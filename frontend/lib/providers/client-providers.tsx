"use client";

import { Toaster } from "@/components/shadcn/sonner";
import { TooltipProvider } from "@/components/shadcn/tooltip";
import { setAccessToken, setAuthLoaded, setUserInfo } from "@/lib/query/client-requestor";
import { getQueryClient } from "@/lib/query/get-query-client";
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { useAuth, useAccessToken } from "@workos-inc/authkit-nextjs/components";
import { useEffect } from "react";

/**
 * Captures the WorkOS access token and user info and makes them available globally
 * for the client-side requestor.
 */
function AuthProvider({ children }: { children: React.ReactNode }) {
  const { accessToken, loading } = useAccessToken();
  const { user } = useAuth();

  useEffect(() => {
    setAccessToken(accessToken);
    if (!loading) {
      setAuthLoaded();
    }
  }, [accessToken, loading]);

  useEffect(() => {
    if (user) {
      setUserInfo({
        email: user.email,
        name: user.firstName && user.lastName
          ? `${user.firstName} ${user.lastName}`
          : user.firstName || user.lastName || "",
      });
    }
  }, [user]);

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
