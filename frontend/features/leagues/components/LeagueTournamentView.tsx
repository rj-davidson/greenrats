"use client";

import type { League } from "../types";
import { LeaguePicksTable } from "./LeaguePicksTable";
import { useBreadcrumbs } from "@/components/core/breadcrumbs";
import { Badge } from "@/components/shadcn/badge";
import { Skeleton } from "@/components/shadcn/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/shadcn/tabs";
import { useLeaguePicks } from "@/features/picks/queries";
import { useTournament } from "@/features/tournaments/queries";
import { CalendarIcon } from "lucide-react";
import Link from "next/link";
import { useEffect } from "react";

interface LeagueTournamentViewProps {
  leagueId: string;
  tournamentId: string;
  league?: League;
}

function formatDateRange(startDate: string, endDate: string): string {
  const start = new Date(startDate);
  const end = new Date(endDate);
  const options: Intl.DateTimeFormatOptions = { month: "short", day: "numeric", year: "numeric" };

  return `${start.toLocaleDateString("en-US", options)} - ${end.toLocaleDateString("en-US", options)}`;
}

function getStatusBadge(status: string) {
  switch (status) {
    case "in_progress":
      return (
        <Badge variant="destructive" className="animate-pulse">
          LIVE
        </Badge>
      );
    case "completed":
      return <Badge variant="secondary">Completed</Badge>;
    case "upcoming":
      return <Badge variant="outline">Upcoming</Badge>;
    default:
      return <Badge variant="outline">{status}</Badge>;
  }
}

export function LeagueTournamentView({
  leagueId,
  tournamentId,
  league,
}: LeagueTournamentViewProps) {
  const { data: tournamentData, isLoading: tournamentLoading } = useTournament(tournamentId);
  const { data: picksData, isLoading: picksLoading } = useLeaguePicks(leagueId, tournamentId);
  const { setExtraCrumbs } = useBreadcrumbs();

  const leagueName = league?.name?.trim();
  const tournamentName = tournamentData?.tournament?.name?.trim();

  useEffect(() => {
    const crumbs: { name: string; path?: string }[] = [];

    if (leagueName) {
      crumbs.push({ name: leagueName, path: `/leagues/${leagueId}` });
    }

    if (tournamentName) {
      crumbs.push({ name: tournamentName });
    }

    setExtraCrumbs(crumbs);
  }, [leagueId, leagueName, setExtraCrumbs, tournamentName]);

  useEffect(() => {
    return () => {
      setExtraCrumbs([]);
    };
  }, [setExtraCrumbs]);

  if (tournamentLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-20 w-full" />
        <Skeleton className="h-96 w-full" />
      </div>
    );
  }

  if (!tournamentData?.tournament) {
    return (
      <div className="text-center">
        <h1 className="mb-2 text-2xl font-bold">Tournament Not Found</h1>
        <p className="text-muted-foreground">The tournament you are looking for does not exist.</p>
      </div>
    );
  }

  const tournament = tournamentData.tournament;

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <Link
            href={`/leagues/${leagueId}`}
            className="text-muted-foreground hover:text-foreground mb-1 block text-sm"
          >
            {league?.name || "Back to league"}
          </Link>
          <h1 className="text-2xl font-bold">{tournament.name}</h1>
          <div className="text-muted-foreground mt-1 flex items-center gap-2">
            <CalendarIcon className="size-4" />
            {formatDateRange(tournament.start_date, tournament.end_date)}
          </div>
        </div>
        {getStatusBadge(tournament.status)}
      </div>

      <Tabs defaultValue="picks" className="space-y-4">
        <TabsList>
          <TabsTrigger value="picks">Picks</TabsTrigger>
          <TabsTrigger value="leaderboard">Leaderboard</TabsTrigger>
        </TabsList>

        <TabsContent value="picks">
          {picksLoading ? (
            <Skeleton className="h-64 w-full" />
          ) : (
            <LeaguePicksTable picks={picksData?.picks ?? []} tournamentStatus={tournament.status} />
          )}
        </TabsContent>

        <TabsContent value="leaderboard">
          <div className="rounded-lg border border-dashed p-12 text-center">
            <p className="text-muted-foreground">Full tournament leaderboard coming soon</p>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
