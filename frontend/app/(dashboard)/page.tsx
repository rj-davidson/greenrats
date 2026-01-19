import { DashboardView } from "@/features/dashboard/components";
import { buildGetUserLeaguesQueryOptions } from "@/features/leagues/queries";
import {
  buildGetActiveTournamentQueryOptions,
  buildGetTournamentsQueryOptions,
} from "@/features/tournaments/queries";
import { buildGetPendingActionsQueryOptions } from "@/features/users/queries";
import { makeServerRequest } from "@/lib/query/server-requestor";
import { HydrationBoundary, QueryClient, dehydrate } from "@tanstack/react-query";

export default async function DashboardPage() {
  const queryClient = new QueryClient();
  await Promise.all([
    queryClient.prefetchQuery(buildGetUserLeaguesQueryOptions(makeServerRequest)),
    queryClient.prefetchQuery(buildGetPendingActionsQueryOptions(makeServerRequest)),
    queryClient.prefetchQuery(buildGetActiveTournamentQueryOptions(makeServerRequest)),
    queryClient.prefetchQuery(
      buildGetTournamentsQueryOptions({ status: "upcoming", limit: 6 }, makeServerRequest),
    ),
  ]);

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <DashboardView />
    </HydrationBoundary>
  );
}
