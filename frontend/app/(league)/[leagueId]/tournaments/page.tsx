"use client";

import { useBreadcrumbs } from "@/components/core/breadcrumbs";
import { Skeleton } from "@/components/shadcn/skeleton";
import { TournamentDataTable, TournamentSpotlightCards } from "@/features/leagues/components";
import { useLeague, useLeagueTournaments } from "@/features/leagues/queries";
import { useParams } from "next/navigation";
import { useEffect } from "react";

export default function TournamentsPage() {
  const params = useParams<{ leagueId: string }>();
  const leagueId = params.leagueId;

  const { data: leagueData } = useLeague(leagueId);
  const { data: tournamentsData, isLoading, error } = useLeagueTournaments(leagueId);
  const { setExtraCrumbs } = useBreadcrumbs();

  const league = leagueData?.league;
  const tournaments = tournamentsData?.tournaments ?? [];

  useEffect(() => {
    if (league?.name) {
      setExtraCrumbs([{ name: league.name, path: `/${leagueId}` }]);
    }
    return () => setExtraCrumbs([]);
  }, [league?.name, leagueId, setExtraCrumbs]);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Tournaments</h1>
        {league && (
          <p className="text-muted-foreground">
            {league.name} &middot; Season {league.season_year}
          </p>
        )}
      </div>

      {isLoading ? (
        <div className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <Skeleton className="h-40" />
            <Skeleton className="h-40" />
            <Skeleton className="h-40" />
          </div>
          <Skeleton className="h-64" />
        </div>
      ) : error ? (
        <div className="text-destructive">Failed to load tournaments</div>
      ) : tournaments.length === 0 ? (
        <div className="py-8 text-center text-muted-foreground">No tournaments found</div>
      ) : (
        <>
          <TournamentSpotlightCards tournaments={tournaments} leagueId={leagueId} />
          <TournamentDataTable tournaments={tournaments} leagueId={leagueId} />
        </>
      )}
    </div>
  );
}
