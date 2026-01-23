"use client";

import { useBreadcrumbs } from "@/components/core/breadcrumbs";
import { Badge } from "@/components/shadcn/badge";
import { Skeleton } from "@/components/shadcn/skeleton";
import { useLeague } from "@/features/leagues/queries";
import { useLeaguePicks } from "@/features/picks/queries";
import { TournamentSelector } from "@/features/tournaments/components";
import { ExpandableLeaderboardTable } from "@/features/tournaments/components/ExpandableLeaderboardTable";
import { useTournament } from "@/features/tournaments/queries";
import { useCurrentUser } from "@/features/users/queries";
import { CalendarIcon } from "lucide-react";
import { useParams } from "next/navigation";
import { useEffect, useMemo } from "react";

function formatDateRange(startDate: string, endDate: string): string {
  const start = new Date(startDate);
  const end = new Date(endDate);
  const options: Intl.DateTimeFormatOptions = { month: "short", day: "numeric", year: "numeric" };

  return `${start.toLocaleDateString("en-US", options)} - ${end.toLocaleDateString("en-US", options)}`;
}

function getStatusBadge(status: string) {
  switch (status) {
    case "active":
      return (
        <Badge variant="destructive" className="animate-pulse">
          LIVE
        </Badge>
      );
    case "completed":
      return <Badge variant="secondary">Final</Badge>;
    case "upcoming":
      return <Badge variant="outline">Upcoming</Badge>;
    default:
      return <Badge variant="outline">{status}</Badge>;
  }
}

export default function TournamentDetailPage() {
  const params = useParams<{ leagueId: string; tournamentId: string }>();
  const { leagueId, tournamentId } = params;

  const { data: leagueData } = useLeague(leagueId);
  const { data: tournamentData, isLoading: tournamentLoading } = useTournament(tournamentId);
  const { data: picksData } = useLeaguePicks(leagueId, tournamentId);
  const { data: currentUser } = useCurrentUser();
  const { setExtraCrumbs } = useBreadcrumbs();

  const league = leagueData?.league;
  const tournament = tournamentData?.tournament;

  const userPickedGolferId = useMemo(() => {
    if (!currentUser || !picksData?.entries) return undefined;
    const entry = picksData.entries.find((p) => p.user_id === currentUser.id);
    return entry?.golfer_id;
  }, [currentUser, picksData]);

  useEffect(() => {
    const crumbs: { name: string; path?: string }[] = [];
    if (league?.name) {
      crumbs.push({ name: league.name, path: `/${leagueId}` });
    }
    if (tournament?.name) {
      crumbs.push({ name: tournament.name });
    }
    setExtraCrumbs(crumbs);
    return () => setExtraCrumbs([]);
  }, [league?.name, tournament?.name, leagueId, setExtraCrumbs]);

  if (tournamentLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-20 w-full" />
        <Skeleton className="h-96 w-full" />
      </div>
    );
  }

  if (!tournament) {
    return (
      <div className="text-center">
        <h1 className="mb-2 text-2xl font-bold">Tournament Not Found</h1>
        <p className="text-muted-foreground">The tournament you are looking for does not exist.</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1">
          <div className="mb-2">
            <TournamentSelector leagueId={leagueId} currentTournamentId={tournamentId} />
          </div>
          <h1 className="text-2xl font-bold">{tournament.name}</h1>
          <div className="mt-1 flex items-center gap-2 text-muted-foreground">
            <CalendarIcon className="size-4" />
            {formatDateRange(tournament.start_date, tournament.end_date)}
          </div>
        </div>
        {getStatusBadge(tournament.status)}
      </div>

      <ExpandableLeaderboardTable
        tournamentId={tournamentId}
        highlightedGolferId={userPickedGolferId}
      />
    </div>
  );
}
