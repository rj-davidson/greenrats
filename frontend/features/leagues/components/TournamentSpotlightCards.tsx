"use client";

import { Badge } from "@/components/shadcn/badge";
import { Button } from "@/components/shadcn/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/shadcn/card";
import { getTournamentSpotlight } from "@/features/leagues/components/tournament-spotlight-utils";
import type { LeagueTournament } from "@/features/leagues/types";
import { getPickWindowState } from "@/features/picks/utils";
import { cn } from "@/lib/utils";
import { CalendarIcon, CheckCircle2Icon, ClockIcon, TrophyIcon, XCircleIcon } from "lucide-react";
import Link from "next/link";

interface TournamentSpotlightCardsProps {
  tournaments: LeagueTournament[];
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

function formatEarnings(earnings: number): string {
  if (earnings >= 1_000_000) {
    return `$${(earnings / 1_000_000).toFixed(2)}M`;
  }
  if (earnings >= 1_000) {
    return `$${(earnings / 1_000).toFixed(0)}K`;
  }
  return `$${earnings.toLocaleString()}`;
}

function formatCountdown(isoString: string): string {
  const target = new Date(isoString);
  const now = new Date();
  const diffMs = target.getTime() - now.getTime();

  if (diffMs <= 0) return "0d 0h 0m";

  const days = Math.floor(diffMs / (1000 * 60 * 60 * 24));
  const hours = Math.floor((diffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
  const minutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));

  return `${days}d ${hours}h ${minutes}m`;
}

type CardVariant = "completed" | "active" | "upcoming";

function getVariant(tournament: LeagueTournament): CardVariant {
  if (tournament.status === "completed") return "completed";
  if (tournament.status === "active") return "active";
  return "upcoming";
}

function getStatusBadge(variant: CardVariant) {
  if (variant === "completed") {
    return <Badge variant="secondary">Final</Badge>;
  }
  return null;
}

interface SpotlightCardProps {
  tournament: LeagueTournament;
  leagueId: string;
  isCenter: boolean;
}

function PickStatus({
  tournament,
  pickWindowState,
}: {
  tournament: LeagueTournament;
  pickWindowState: "not_open" | "open" | "closed";
}) {
  const hasEarnings =
    tournament.status === "completed" &&
    tournament.golfer_earnings !== undefined &&
    tournament.golfer_earnings > 0;

  if (tournament.has_user_pick) {
    return (
      <div className="space-y-1">
        <div className="flex items-center gap-1.5 text-sm text-green-600">
          <CheckCircle2Icon className="size-4 shrink-0" />
          <span className="truncate">{tournament.golfer_name}</span>
        </div>
        {hasEarnings && (
          <div className="flex items-center gap-1.5 text-sm">
            <TrophyIcon className="size-4 shrink-0 text-yellow-500" />
            <span className="font-medium">{formatEarnings(tournament.golfer_earnings!)}</span>
          </div>
        )}
      </div>
    );
  }

  if (pickWindowState === "not_open" && tournament.pick_window_opens_at) {
    return (
      <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
        <ClockIcon className="size-4 shrink-0" />
        <span>Picks open in {formatCountdown(tournament.pick_window_opens_at)}</span>
      </div>
    );
  }

  if (pickWindowState === "open") {
    return (
      <div className="space-y-2">
        {tournament.pick_window_closes_at && (
          <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
            <ClockIcon className="size-4 shrink-0" />
            <span>Picks close in {formatCountdown(tournament.pick_window_closes_at)}</span>
          </div>
        )}
        <Button size="sm" className="w-full">
          Make Pick
        </Button>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-1.5 text-sm">
      <XCircleIcon className="size-4 shrink-0" />
      <span>Did not pick</span>
    </div>
  );
}

function SpotlightCard({ tournament, leagueId, isCenter }: SpotlightCardProps) {
  const variant = getVariant(tournament);
  const pickWindowState = getPickWindowState(tournament);

  return (
    <Link href={`/${leagueId}/tournaments/${tournament.id}`} className="block">
      <Card
        className={cn(
          "h-full transition-colors hover:bg-muted/50",
          isCenter && "border-2 border-primary/50 shadow-md",
        )}
      >
        <CardHeader className="flex flex-row items-start justify-between gap-2 pb-2">
          <CardTitle className="line-clamp-2 text-base font-semibold">{tournament.name}</CardTitle>
          {getStatusBadge(variant)}
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <CalendarIcon className="size-4 shrink-0" />
            {formatDateRange(tournament.start_date, tournament.end_date)}
          </div>
          <PickStatus tournament={tournament} pickWindowState={pickWindowState} />
        </CardContent>
      </Card>
    </Link>
  );
}

export function TournamentSpotlightCards({ tournaments, leagueId }: TournamentSpotlightCardsProps) {
  const spotlightTournaments = getTournamentSpotlight(tournaments);

  if (spotlightTournaments.length === 0) {
    return null;
  }

  const centerIndex = Math.floor(spotlightTournaments.length / 2);

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {spotlightTournaments.map((tournament, index) => (
        <SpotlightCard
          key={tournament.id}
          tournament={tournament}
          leagueId={leagueId}
          isCenter={index === centerIndex && spotlightTournaments.length === 3}
        />
      ))}
    </div>
  );
}
