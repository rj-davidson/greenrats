"use client";

import { useBreadcrumbs } from "@/components/core/breadcrumbs";
import { CommissionerPanel, LeagueMonogram } from "@/features/leagues/components";
import { useLeague } from "@/features/leagues/queries";
import { notFound, useParams } from "next/navigation";
import { useEffect } from "react";

export default function ManagePage() {
  const params = useParams<{ leagueId: string }>();
  const leagueId = params.leagueId;

  const { data: leagueData, isLoading } = useLeague(leagueId);
  const { setExtraCrumbs } = useBreadcrumbs();

  const league = leagueData?.league;
  const isOwner = league?.role === "owner";

  useEffect(() => {
    if (league?.name) {
      setExtraCrumbs([{ name: league.name, path: `/${leagueId}` }]);
    }
    return () => setExtraCrumbs([]);
  }, [league?.name, leagueId, setExtraCrumbs]);

  if (!isLoading && league && !isOwner) {
    notFound();
  }

  if (!league) {
    return null;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <LeagueMonogram league={league} size={40} />
        <div>
          <h1 className="text-2xl font-bold">Manage League</h1>
          <p className="text-muted-foreground">Commissioner settings for {league.name}</p>
        </div>
      </div>

      <CommissionerPanel league={league} />
    </div>
  );
}
