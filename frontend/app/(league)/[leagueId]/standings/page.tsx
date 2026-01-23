"use client";

import { useBreadcrumbs } from "@/components/core/breadcrumbs";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/shadcn/card";
import { LeagueLeaderboard } from "@/features/leaderboards/components";
import { LeagueMonogram } from "@/features/leagues/components";
import { useLeague } from "@/features/leagues/queries";
import { useParams } from "next/navigation";
import { useEffect } from "react";

export default function StandingsPage() {
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
      <div className="flex items-center gap-4">
        {league && <LeagueMonogram league={league} size={40} />}
        <div>
          <h1 className="text-2xl font-bold">Season Standings</h1>
          {league && (
            <p className="text-muted-foreground">
              {league.name} &middot; Season {league.season_year}
            </p>
          )}
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Leaderboard</CardTitle>
        </CardHeader>
        <CardContent>
          <LeagueLeaderboard leagueId={leagueId} />
        </CardContent>
      </Card>
    </div>
  );
}
