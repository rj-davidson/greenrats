"use client";

import { useBreadcrumbs } from "@/components/core/breadcrumbs";
import { LeagueTournamentList } from "@/features/leagues/components";
import { useLeague } from "@/features/leagues/queries";
import { TournamentSelector } from "@/features/tournaments/components";
import { useParams } from "next/navigation";
import { useEffect } from "react";

export default function TournamentsPage() {
  const params = useParams<{ leagueId: string }>();
  const leagueId = params.leagueId;

  const { data: leagueData } = useLeague(leagueId);
  const { setExtraCrumbs } = useBreadcrumbs();

  const league = leagueData?.league;

  useEffect(() => {
    if (league?.name) {
      setExtraCrumbs([{ name: league.name, path: `/${leagueId}` }]);
    }
    return () => setExtraCrumbs([]);
  }, [league?.name, leagueId, setExtraCrumbs]);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Tournaments</h1>
          {league && (
            <p className="text-muted-foreground">
              {league.name} &middot; Season {league.season_year}
            </p>
          )}
        </div>
        <TournamentSelector leagueId={leagueId} />
      </div>

      <LeagueTournamentList leagueId={leagueId} />
    </div>
  );
}
