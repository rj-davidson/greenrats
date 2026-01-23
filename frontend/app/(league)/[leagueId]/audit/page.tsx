"use client";

import { useBreadcrumbs } from "@/components/core/breadcrumbs";
import { LeagueActivity, LeagueMonogram } from "@/features/leagues/components";
import { useLeague } from "@/features/leagues/queries";
import { useParams } from "next/navigation";
import { useEffect } from "react";

export default function AuditPage() {
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
          <h1 className="text-2xl font-bold">Audit Log</h1>
          {league && (
            <p className="text-muted-foreground">
              Commissioner actions and changes for {league.name}
            </p>
          )}
        </div>
      </div>

      <LeagueActivity leagueId={leagueId} />
    </div>
  );
}
