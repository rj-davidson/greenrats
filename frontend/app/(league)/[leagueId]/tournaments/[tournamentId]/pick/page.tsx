"use client";

import { useBreadcrumbs } from "@/components/core/breadcrumbs";
import { Skeleton } from "@/components/shadcn/skeleton";
import { useLeague } from "@/features/leagues/queries";
import {
  PickFieldTable,
  PickFieldSkeleton,
  TournamentPickHeader,
} from "@/features/picks/components";
import { usePickField } from "@/features/picks/queries";
import { ArrowLeftIcon } from "lucide-react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useMemo } from "react";

export default function PickPage() {
  const params = useParams<{ leagueId: string; tournamentId: string }>();
  const { leagueId, tournamentId } = params;

  const { data: leagueData } = useLeague(leagueId);
  const { data, isLoading, error } = usePickField(leagueId, tournamentId);
  const { setExtraCrumbs } = useBreadcrumbs();

  const league = leagueData?.league;

  const currentPickGolferName = useMemo(() => {
    if (!data?.current_pick_golfer_id) return undefined;
    const entry = data.entries.find((e) => e.golfer_id === data.current_pick_golfer_id);
    return entry?.golfer_name;
  }, [data]);

  useEffect(() => {
    const crumbs: { name: string; path?: string }[] = [];
    if (league?.name) {
      crumbs.push({ name: league.name, path: `/${leagueId}` });
    }
    if (data?.tournament_name) {
      crumbs.push({ name: data.tournament_name, path: `/${leagueId}/tournaments/${tournamentId}` });
      crumbs.push({ name: "Pick" });
    }
    setExtraCrumbs(crumbs);
    return () => setExtraCrumbs([]);
  }, [league?.name, data?.tournament_name, leagueId, tournamentId, setExtraCrumbs]);

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-40 w-full" />
        <PickFieldSkeleton />
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center">
        <h1 className="mb-2 text-2xl font-bold">Error</h1>
        <p className="text-muted-foreground">Failed to load pick field. Please try again.</p>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="text-center">
        <h1 className="mb-2 text-2xl font-bold">Tournament Not Found</h1>
        <p className="text-muted-foreground">The tournament you are looking for does not exist.</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <Link
          href={`/${leagueId}/tournaments/${tournamentId}`}
          className="mb-3 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeftIcon className="size-4" />
          Back to Tournament
        </Link>
      </div>

      <TournamentPickHeader data={data} currentPickGolferName={currentPickGolferName} />

      <PickFieldTable data={data} leagueId={leagueId} />
    </div>
  );
}
