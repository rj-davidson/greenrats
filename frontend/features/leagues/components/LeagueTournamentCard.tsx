"use client";

import type { LeagueTournament } from "@/features/leagues/types";
import { formatPickWindowDate } from "@/features/picks/utils";
import { Badge } from "@/components/shadcn/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/shadcn/card";
import { CalendarIcon, CheckCircle2Icon, ClockIcon, UsersIcon } from "lucide-react";
import Link from "next/link";
import { cn } from "@/lib/utils";

export type TournamentCardVariant = "live" | "next" | "upcoming" | "final";

interface LeagueTournamentCardProps {
  tournament: LeagueTournament;
  leagueId: string;
  variant: TournamentCardVariant;
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

function getStatusBadge(variant: TournamentCardVariant, compact = false) {
  const size = compact ? "text-xs px-1.5 py-0" : "";
  switch (variant) {
    case "live":
      return (
        <Badge variant="destructive" className={cn("animate-pulse", size)}>
          LIVE
        </Badge>
      );
    case "next":
      return (
        <Badge variant="outline" className={cn("border-primary text-primary", size)}>
          Up Next
        </Badge>
      );
    case "upcoming":
      return (
        <Badge variant="outline" className={size}>
          Upcoming
        </Badge>
      );
    case "final":
      return (
        <Badge variant="secondary" className={size}>
          Final
        </Badge>
      );
  }
}

function CompactRow({ tournament, leagueId, variant }: LeagueTournamentCardProps) {
  const isMuted = variant === "upcoming";

  return (
    <Link href={`/${leagueId}/tournaments/${tournament.id}`}>
      <div
        className={cn(
          "flex flex-col gap-1 rounded-md border px-3 py-2 transition-colors hover:bg-muted/50 sm:flex-row sm:items-center sm:gap-3",
          isMuted && "opacity-60",
        )}
      >
        <div className="flex items-center justify-between gap-2 sm:contents">
          <span className="truncate text-sm sm:min-w-0 sm:flex-1">{tournament.name}</span>
          <div className="sm:order-last">{getStatusBadge(variant, true)}</div>
        </div>

        <div className="flex items-center justify-between gap-2 text-xs text-muted-foreground sm:contents">
          <span className="sm:w-28 sm:shrink-0">
            {formatDateRange(tournament.start_date, tournament.end_date)}
          </span>
          <div className="shrink-0 sm:ml-auto">
            {tournament.has_user_pick ? (
              <div className="flex items-center gap-1 text-green-600">
                <CheckCircle2Icon className="size-3" />
                <span>{tournament.golfer_name}</span>
                {variant === "final" && tournament.golfer_earnings !== undefined && tournament.golfer_earnings > 0 && (
                  <span className="font-medium">{formatEarnings(tournament.golfer_earnings)}</span>
                )}
              </div>
            ) : variant === "upcoming" ? (
              <span>No pick</span>
            ) : null}
          </div>
        </div>
      </div>
    </Link>
  );
}

function FullCard({ tournament, leagueId, variant }: LeagueTournamentCardProps) {
  const showPickWindow =
    (variant === "next" || variant === "upcoming") &&
    tournament.pick_window_opens_at &&
    tournament.pick_window_closes_at;

  return (
    <Link href={`/${leagueId}/tournaments/${tournament.id}`}>
      <Card className="border-2 border-primary/50 shadow-sm transition-colors hover:bg-muted/50">
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-base font-semibold">{tournament.name}</CardTitle>
          {getStatusBadge(variant)}
        </CardHeader>
        <CardContent className="space-y-2">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <CalendarIcon className="size-4" />
            {formatDateRange(tournament.start_date, tournament.end_date)}
          </div>
          {showPickWindow && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <ClockIcon className="size-4" />
              <span>
                Picks: {formatPickWindowDate(tournament.pick_window_opens_at!)} -{" "}
                {formatPickWindowDate(tournament.pick_window_closes_at!)}
              </span>
            </div>
          )}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <UsersIcon className="size-4" />
              {tournament.pick_count} {tournament.pick_count === 1 ? "pick" : "picks"}
            </div>
            {tournament.has_user_pick ? (
              <div className="flex items-center gap-1.5 text-sm text-green-600">
                <CheckCircle2Icon className="size-4" />
                <span>{tournament.golfer_name || "Pick made"}</span>
              </div>
            ) : (
              <span className="text-sm text-muted-foreground">No pick yet</span>
            )}
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}

export function LeagueTournamentCard(props: LeagueTournamentCardProps) {
  const isCompact = props.variant === "final" || props.variant === "upcoming";

  if (isCompact) {
    return <CompactRow {...props} />;
  }

  return <FullCard {...props} />;
}
