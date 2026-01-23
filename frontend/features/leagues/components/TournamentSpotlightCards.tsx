"use client";

import { Badge } from "@/components/shadcn/badge";
import { Button } from "@/components/shadcn/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/shadcn/card";
import { getTournamentSpotlight } from "@/features/leagues/components/tournament-spotlight-utils";
import type { LeagueTournament } from "@/features/leagues/types";
import { getPickWindowState } from "@/features/picks/utils";
import { cn } from "@/lib/utils";
import { CalendarIcon, CheckCircle2Icon, CircleIcon, TrophyIcon } from "lucide-react";
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

type CardVariant = "completed" | "active" | "upcoming";

function getVariant(tournament: LeagueTournament): CardVariant {
  if (tournament.status === "completed") return "completed";
  if (tournament.status === "active") return "active";
  return "upcoming";
}

function getStatusBadge(variant: CardVariant) {
  switch (variant) {
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
  }
}

interface SpotlightCardProps {
  tournament: LeagueTournament;
  leagueId: string;
  isCenter: boolean;
}

function SpotlightCard({ tournament, leagueId, isCenter }: SpotlightCardProps) {
  const variant = getVariant(tournament);
  const pickWindowState = getPickWindowState(tournament);
  const canMakePick = !tournament.has_user_pick && pickWindowState === "open";
  const hasEarnings =
    variant === "completed" &&
    tournament.golfer_earnings !== undefined &&
    tournament.golfer_earnings > 0;

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

          {tournament.has_user_pick ? (
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
          ) : (
            <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
              <CircleIcon className="size-4 shrink-0" />
              <span>No pick</span>
            </div>
          )}

          {canMakePick && (
            <Button size="sm" className="w-full">
              Make Pick
            </Button>
          )}
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
