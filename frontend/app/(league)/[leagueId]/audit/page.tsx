"use client";

import { useBreadcrumbs } from "@/components/core/breadcrumbs";
import { Button } from "@/components/shadcn/button";
import { LeagueActivity } from "@/features/leagues/components";
import { useLeague } from "@/features/leagues/queries";
import { ArrowLeftIcon } from "lucide-react";
import Link from "next/link";
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
      <div className="space-y-4">
        <Button variant="ghost" size="sm" asChild>
          <Link href={`/${leagueId}`}>
            <ArrowLeftIcon className="size-4" />
            Back to League
          </Link>
        </Button>
        <div>
          <h1 className="text-2xl font-bold">Audit</h1>
          <p className="text-muted-foreground">
            A record of pick changes and league setting updates made by commissioners.
          </p>
        </div>
      </div>

      <LeagueActivity leagueId={leagueId} />
    </div>
  );
}
