"use client";

import { Toaster } from "@/components/shadcn/sonner";
import { TooltipProvider } from "@/components/shadcn/tooltip";
import { buildGetCurrentUserQueryOptions } from "@/features/users/queries";
import { setAccessToken, setAuthLoaded, setUserInfo } from "@/lib/query/client-requestor";
import { getQueryClient } from "@/lib/query/get-query-client";
import * as Sentry from "@sentry/nextjs";
import { useQuery, QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { useAuth, useAccessToken } from "@workos-inc/authkit-nextjs/components";
import { ThemeProvider } from "next-themes";
import { useEffect } from "react";

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
        name:
          user.firstName && user.lastName
            ? `${user.firstName} ${user.lastName}`
            : user.firstName || user.lastName || "",
      });
    }
  }, [user]);

  return <>{children}</>;
}

function QueryProvider({ children }: { children: React.ReactNode }) {
  const queryClient = getQueryClient();

  return (
    <QueryClientProvider client={queryClient}>
      {children}
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
}

function ObservabilityProvider({ children }: { children: React.ReactNode }) {
  const { user: workosUser } = useAuth();
  const { data: dbUser } = useQuery({
    ...buildGetCurrentUserQueryOptions(),
    enabled: !!workosUser,
  });

  useEffect(() => {
    if (dbUser) {
      Sentry.setUser({
        id: dbUser.id,
        email: dbUser.email,
        username: dbUser.display_name ?? undefined,
      });
    } else if (!workosUser) {
      Sentry.setUser(null);
    }
  }, [dbUser, workosUser]);

  return <>{children}</>;
}

export function ClientProviders({ children }: { children: React.ReactNode }) {
  return (
    <AuthProvider>
      <QueryProvider>
        <ObservabilityProvider>
          <ThemeProvider
            attribute="class"
            defaultTheme="system"
            enableSystem
            disableTransitionOnChange
          >
            <TooltipProvider>
              {children}
              <Toaster />
            </TooltipProvider>
          </ThemeProvider>
        </ObservabilityProvider>
      </QueryProvider>
    </AuthProvider>
  );
}
