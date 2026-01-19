import { DashboardView } from "@/features/dashboard/components";
import { buildGetUserLeaguesQueryOptions } from "@/features/leagues/queries";
import { buildGetPendingActionsQueryOptions } from "@/features/users/queries";
import { makeServerRequest } from "@/lib/query/server-requestor";
import { HydrationBoundary, QueryClient, dehydrate } from "@tanstack/react-query";

export default async function DashboardPage() {
  const queryClient = new QueryClient();
  await Promise.all([
    queryClient.prefetchQuery(buildGetUserLeaguesQueryOptions(makeServerRequest)),
    queryClient.prefetchQuery(buildGetPendingActionsQueryOptions(makeServerRequest)),
  ]);

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <DashboardView />
    </HydrationBoundary>
  );
}
