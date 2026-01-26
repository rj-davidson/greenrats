"use client";

import { useBreadcrumbs } from "@/components/core/breadcrumbs";
import { LeagueActivity } from "@/features/leagues/components";
import { useParams } from "next/navigation";
import { useEffect } from "react";

export default function AuditPage() {
  const params = useParams<{ leagueId: string }>();
  const leagueId = params.leagueId;

  const { setExtraCrumbs } = useBreadcrumbs();

  useEffect(() => {
    setExtraCrumbs([{ name: "Audit Log" }]);
    return () => setExtraCrumbs([]);
  }, [setExtraCrumbs]);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Audit</h1>
        <p className="text-muted-foreground">
          A record of pick changes and league setting updates made by commissioners.
        </p>
      </div>

      <LeagueActivity leagueId={leagueId} />
    </div>
  );
}
