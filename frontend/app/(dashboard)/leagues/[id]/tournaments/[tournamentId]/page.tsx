import { LeagueTournamentView } from "@/features/leagues/components";
import { buildGetLeagueQueryOptions } from "@/features/leagues/queries";
import { buildGetLeaguePicksQueryOptions } from "@/features/picks/queries";
import { buildGetTournamentQueryOptions } from "@/features/tournaments/queries";
import { makeServerRequest } from "@/lib/query/server-requestor";
import { HydrationBoundary, QueryClient, dehydrate } from "@tanstack/react-query";

interface LeagueTournamentPageProps {
  params: Promise<{ id: string; tournamentId: string }>;
}

export default async function LeagueTournamentPage({ params }: LeagueTournamentPageProps) {
  const { id: leagueId, tournamentId } = await params;

  const queryClient = new QueryClient();
  await Promise.all([
    queryClient.prefetchQuery(buildGetLeagueQueryOptions(leagueId, makeServerRequest)),
    queryClient.prefetchQuery(buildGetTournamentQueryOptions(tournamentId, makeServerRequest)),
    queryClient.prefetchQuery(
      buildGetLeaguePicksQueryOptions(leagueId, tournamentId, makeServerRequest),
    ),
  ]);

  const leagueData = queryClient.getQueryData<{ league: { name: string } }>(
    buildGetLeagueQueryOptions(leagueId).queryKey,
  );

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <div className="container mx-auto p-8">
        <LeagueTournamentView
          leagueId={leagueId}
          tournamentId={tournamentId}
          league={leagueData?.league as any}
        />
      </div>
    </HydrationBoundary>
  );
}
