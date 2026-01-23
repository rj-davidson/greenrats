"use client";

import { DashboardCard } from "./DashboardCard";
import { Badge } from "@/components/shadcn/badge";
import { Button } from "@/components/shadcn/button";
import { useLeagueTournaments } from "@/features/leagues/queries";
import { useLeaderboard, useTournament } from "@/features/tournaments/queries";
import { CalendarIcon, DollarSignIcon, FlagIcon, MapPinIcon, TrophyIcon } from "lucide-react";
import Link from "next/link";
import { useMemo } from "react";

interface ActiveTournamentCardProps {
  leagueId: string;
}

function formatDateRange(startDate: string, endDate: string): string {
  const start = new Date(startDate);
  const end = new Date(endDate);
  const options: Intl.DateTimeFormatOptions = { month: "short", day: "numeric" };

  if (start.getMonth() === end.getMonth()) {
    return `${start.toLocaleDateString("en-US", options)} - ${end.getDate()}`;
  }
  return `${start.toLocaleDateString("en-US", options)} - ${end.toLocaleDateString("en-US", options)}`;
}

function formatPurse(purse: number): string {
  if (purse >= 1_000_000) {
    return `$${(purse / 1_000_000).toFixed(0)}M`;
  }
  if (purse >= 1_000) {
    return `$${(purse / 1_000).toFixed(0)}K`;
  }
  return `$${purse.toLocaleString()}`;
}

function formatLocation(city?: string, state?: string, country?: string): string | null {
  if (city && state) return `${city}, ${state}`;
  if (city && country) return `${city}, ${country}`;
  if (city) return city;
  if (state) return state;
  if (country) return country;
  return null;
}

function getRoundLabel(roundNumber: number): string {
  if (roundNumber === 4) return "Final Round";
  if (roundNumber === 1) return "Round 1";
  if (roundNumber === 2) return "Round 2";
  if (roundNumber === 3) return "Round 3";
  return `Round ${roundNumber}`;
}

export function ActiveTournamentCard({ leagueId }: ActiveTournamentCardProps) {
  const { data, isLoading } = useLeagueTournaments(leagueId);

  const activeTournamentId = useMemo(() => {
    if (!data?.tournaments) return null;
    const active = data.tournaments.find((t) => t.status === "active");
    return active?.id ?? null;
  }, [data]);

  const { data: tournamentData, isLoading: tournamentLoading } = useTournament(
    activeTournamentId ?? "",
  );
  const { data: leaderboardData } = useLeaderboard(activeTournamentId ?? "");

  const tournament = tournamentData?.tournament;
  const currentRound = leaderboardData?.current_round;

  if (isLoading || tournamentLoading) {
    return (
      <DashboardCard title="Live Tournament" icon={<TrophyIcon className="size-4" />} isLoading />
    );
  }

  if (!activeTournamentId || !tournament) {
    return null;
  }

  const location = formatLocation(tournament.city, tournament.state, tournament.country);

  const action = (
    <Button asChild variant="ghost" size="sm">
      <Link href={`/${leagueId}/tournaments/${tournament.id}`}>View</Link>
    </Button>
  );

  return (
    <DashboardCard title="Live Tournament" icon={<TrophyIcon className="size-4" />} action={action}>
      <div className="space-y-4">
        <div>
          <div className="mb-1 flex items-center gap-2">
            <Badge variant="default" className="text-xs">
              {currentRound ? getRoundLabel(currentRound) : "Live"}
            </Badge>
          </div>
          <h3 className="text-lg font-semibold">{tournament.name}</h3>
        </div>

        <div className="space-y-2 text-sm text-muted-foreground">
          {tournament.course && (
            <div className="flex items-center gap-2">
              <FlagIcon className="size-4 shrink-0" />
              <span>{tournament.course}</span>
            </div>
          )}

          {location && (
            <div className="flex items-center gap-2">
              <MapPinIcon className="size-4 shrink-0" />
              <span>{location}</span>
            </div>
          )}

          <div className="flex items-center gap-2">
            <CalendarIcon className="size-4 shrink-0" />
            <span>{formatDateRange(tournament.start_date, tournament.end_date)}</span>
          </div>

          {tournament.purse && (
            <div className="flex items-center gap-2">
              <DollarSignIcon className="size-4 shrink-0" />
              <span>{formatPurse(tournament.purse)} purse</span>
            </div>
          )}
        </div>
      </div>
    </DashboardCard>
  );
}
