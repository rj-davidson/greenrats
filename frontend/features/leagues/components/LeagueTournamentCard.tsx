"use client";

import type { LeagueTournament } from "@/features/leagues/types";
import { Badge } from "@/components/shadcn/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/shadcn/card";
import { CalendarIcon, CheckCircle2Icon, UsersIcon } from "lucide-react";
import Link from "next/link";

interface LeagueTournamentCardProps {
  tournament: LeagueTournament;
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

export function LeagueTournamentCard({ tournament, leagueId }: LeagueTournamentCardProps) {
  return (
    <Link href={`/leagues/${leagueId}/tournaments/${tournament.id}`}>
      <Card className="transition-colors hover:bg-muted/50">
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-base font-medium">{tournament.name}</CardTitle>
          {getStatusBadge(tournament.status)}
        </CardHeader>
        <CardContent className="space-y-2">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <CalendarIcon className="size-4" />
            {formatDateRange(tournament.start_date, tournament.end_date)}
          </div>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <UsersIcon className="size-4" />
              {tournament.pick_count} {tournament.pick_count === 1 ? "pick" : "picks"}
            </div>
            {tournament.has_user_pick ? (
              <div className="flex items-center gap-1.5 text-sm text-green-600">
                <CheckCircle2Icon className="size-4" />
                {tournament.golfer_name || "Pick made"}
              </div>
            ) : tournament.status === "upcoming" ? (
              <span className="text-sm text-muted-foreground">No pick yet</span>
            ) : null}
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}
