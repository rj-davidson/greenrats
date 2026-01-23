import { Breadcrumbs, BreadcrumbsProvider } from "@/components/core/breadcrumbs";
import { LeagueSidebar } from "@/components/core/league-sidebar";
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/shadcn/sidebar";
import { buildGetLeagueQueryOptions } from "@/features/leagues/queries";
import type { User } from "@/features/users/types";
import { makeServerRequest } from "@/lib/query/server-requestor";
import { HydrationBoundary, QueryClient, dehydrate } from "@tanstack/react-query";
import { notFound, redirect } from "next/navigation";

interface LeagueLayoutProps {
  children: React.ReactNode;
  params: Promise<{ leagueId: string }>;
}

export default async function LeagueLayout({ children, params }: LeagueLayoutProps) {
  const { leagueId } = await params;

  let user: User;
  try {
    user = await makeServerRequest.get<User>("/api/v1/users/me");
  } catch {
    redirect("/login");
  }

  if (!user.display_name) {
    redirect("/onboarding");
  }

  const queryClient = new QueryClient();

  try {
    await queryClient.prefetchQuery(buildGetLeagueQueryOptions(leagueId, makeServerRequest));
  } catch {
    notFound();
  }

  const leagueData = queryClient.getQueryData(buildGetLeagueQueryOptions(leagueId).queryKey);

  if (!leagueData) {
    notFound();
  }

  const league = (leagueData as { league: { id: string; name: string; role?: string } }).league;

  if (!league.role) {
    notFound();
  }

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <SidebarProvider>
        <BreadcrumbsProvider>
          <LeagueSidebar
            league={league as { id: string; name: string; role: "owner" | "member" }}
          />
          <SidebarInset>
            <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
              <SidebarTrigger className="-ml-1" />
              <Breadcrumbs />
            </header>
            <main className="min-w-0 flex-1 overflow-x-hidden p-4">{children}</main>
          </SidebarInset>
        </BreadcrumbsProvider>
      </SidebarProvider>
    </HydrationBoundary>
  );
}
