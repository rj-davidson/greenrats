import { AppSidebar } from "@/components/core/app-sidebar";
import { Breadcrumbs, BreadcrumbsProvider } from "@/components/core/breadcrumbs";
import { Button } from "@/components/shadcn/button";
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/shadcn/sidebar";
import { DashboardView } from "@/features/dashboard/components";
import { buildGetUserLeaguesQueryOptions } from "@/features/leagues/queries";
import {
  buildGetActiveTournamentQueryOptions,
  buildGetTournamentsQueryOptions,
} from "@/features/tournaments/queries";
import { buildGetPendingActionsQueryOptions } from "@/features/users/queries";
import type { User } from "@/features/users/types";
import { makeServerRequest } from "@/lib/query/server-requestor";
import { DehydratedState, HydrationBoundary, QueryClient, dehydrate } from "@tanstack/react-query";
import { withAuth } from "@workos-inc/authkit-nextjs";
import Link from "next/link";
import { redirect } from "next/navigation";

function LandingPage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-8">
      <div className="max-w-2xl text-center">
        <h1 className="mb-4 text-5xl font-bold">GreenRats</h1>
        <p className="text-muted-foreground mb-8 text-xl">
          Pick one golfer per tournament. Compete with friends. Track your earnings throughout the
          PGA Tour season.
        </p>
        <div className="flex justify-center gap-4">
          <Link href="/login">
            <Button size="lg">Get Started</Button>
          </Link>
          <Link href="/rules">
            <Button variant="outline" size="lg">
              How to Play
            </Button>
          </Link>
        </div>
      </div>
    </main>
  );
}

interface DashboardHomeProps {
  dehydratedState: DehydratedState;
}

function DashboardHome({ dehydratedState }: DashboardHomeProps) {
  return (
    <SidebarProvider>
      <BreadcrumbsProvider>
        <AppSidebar />
        <SidebarInset>
          <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
            <SidebarTrigger className="-ml-1" />
            <Breadcrumbs />
          </header>
          <main className="min-w-0 flex-1 overflow-x-hidden p-4">
            <HydrationBoundary state={dehydratedState}>
              <DashboardView />
            </HydrationBoundary>
          </main>
        </SidebarInset>
      </BreadcrumbsProvider>
    </SidebarProvider>
  );
}

export default async function Home() {
  const { user: authUser } = await withAuth();

  if (!authUser) {
    return <LandingPage />;
  }

  let user: User | null = null;
  try {
    user = await makeServerRequest.get<User>("/api/v1/users/me");
  } catch (error) {
    console.error("Failed to fetch user:", error);
  }

  if (user && !user.display_name) {
    redirect("/onboarding");
  }

  const queryClient = new QueryClient();
  await Promise.all([
    queryClient.prefetchQuery(buildGetUserLeaguesQueryOptions(makeServerRequest)),
    queryClient.prefetchQuery(buildGetPendingActionsQueryOptions(makeServerRequest)),
    queryClient.prefetchQuery(buildGetActiveTournamentQueryOptions(makeServerRequest)),
    queryClient.prefetchQuery(
      buildGetTournamentsQueryOptions({ status: "upcoming", limit: 6 }, makeServerRequest),
    ),
  ]);

  return <DashboardHome dehydratedState={dehydrate(queryClient)} />;
}
