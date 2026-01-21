"use client";

import { useLeagueTournaments } from "@/features/leagues/queries";
import { LeagueTournamentCard, type TournamentCardVariant } from "@/features/leagues/components/LeagueTournamentCard";
import type { LeagueTournament } from "@/features/leagues/types";
import { Skeleton } from "@/components/shadcn/skeleton";
import { useMemo } from "react";

interface LeagueTournamentListProps {
  leagueId: string;
}

function getVariant(tournament: LeagueTournament, firstUpcomingId: string | null): TournamentCardVariant {
  if (tournament.status === "in_progress") {
    return "live";
  }
  if (tournament.status === "completed") {
    return "final";
  }
  if (tournament.id === firstUpcomingId) {
    return "next";
  }
  return "upcoming";
}

export function LeagueTournamentList({ leagueId }: LeagueTournamentListProps) {
  const { data, isLoading, error } = useLeagueTournaments(leagueId);

  const tournaments = data?.tournaments;
  const sortedTournaments = useMemo(() => {
    if (!tournaments) return [];
    return [...tournaments].sort(
      (a, b) => new Date(a.start_date).getTime() - new Date(b.start_date).getTime(),
    );
  }, [tournaments]);

  const firstUpcomingId = useMemo(() => {
    const firstUpcoming = sortedTournaments.find((t) => t.status === "upcoming");
    return firstUpcoming?.id ?? null;
  }, [sortedTournaments]);

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-32 w-full" />
      </div>
    );
  }

  if (error) {
    return <div className="text-destructive">Failed to load tournaments</div>;
  }

  if (!data || data.tournaments.length === 0) {
    return <div className="py-8 text-center text-muted-foreground">No tournaments found</div>;
  }

  return (
    <div className="flex flex-col gap-4">
      {sortedTournaments.map((tournament) => (
        <LeagueTournamentCard
          key={tournament.id}
          tournament={tournament}
          leagueId={leagueId}
          variant={getVariant(tournament, firstUpcomingId)}
        />
      ))}
    </div>
  );
}
