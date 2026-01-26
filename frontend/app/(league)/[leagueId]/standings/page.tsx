"use client";

import { useBreadcrumbs } from "@/components/core/breadcrumbs";
import { ExpandableLeagueStandings } from "@/features/leaderboards/components";
import { useLeague } from "@/features/leagues/queries";
import { TrophyIcon } from "lucide-react";
import { useParams } from "next/navigation";
import { useEffect } from "react";

export default function StandingsPage() {
  const params = useParams<{ leagueId: string }>();
  const leagueId = params.leagueId;

  const { data: leagueData } = useLeague(leagueId);
  const { setExtraCrumbs } = useBreadcrumbs();

  const league = leagueData?.league;

  useEffect(() => {
    setExtraCrumbs([{ name: "Standings" }]);
    return () => setExtraCrumbs([]);
  }, [setExtraCrumbs]);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Season Standings</h1>
        {league && (
          <div className="mt-1 flex items-center gap-2 text-muted-foreground">
            <TrophyIcon className="size-4" />
            {league.name} &middot; Season {league.season_year}
          </div>
        )}
      </div>

      <ExpandableLeagueStandings leagueId={leagueId} />
    </div>
  );
}
