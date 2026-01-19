import { LeagueDetail } from "./league-detail";
import { buildGetLeagueQueryOptions } from "@/features/leagues/queries";
import { makeServerRequest } from "@/lib/query/server-requestor";
import { HydrationBoundary, QueryClient, dehydrate } from "@tanstack/react-query";

interface LeaguePageProps {
  params: Promise<{ id: string }>;
}

export default async function LeaguePage({ params }: LeaguePageProps) {
  const { id } = await params;

  const queryClient = new QueryClient();
  await queryClient.prefetchQuery(buildGetLeagueQueryOptions(id, makeServerRequest));

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <LeagueDetail id={id} />
    </HydrationBoundary>
  );
}
