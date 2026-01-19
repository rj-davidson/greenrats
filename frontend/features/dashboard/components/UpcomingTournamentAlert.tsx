"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/shadcn/card";
import { Skeleton } from "@/components/shadcn/skeleton";
import { usePendingActions } from "@/features/users/queries";
import type { UpcomingTournament } from "@/features/users/types";
import { CalendarIcon } from "lucide-react";

function formatDateRange(startDate: string, endDate: string): string {
  const start = new Date(startDate);
  const end = new Date(endDate);
  const options: Intl.DateTimeFormatOptions = { month: "short", day: "numeric" };

  if (start.getMonth() === end.getMonth()) {
    return `${start.toLocaleDateString("en-US", options)} - ${end.getDate()}`;
  }
  return `${start.toLocaleDateString("en-US", options)} - ${end.toLocaleDateString("en-US", options)}`;
}

function TournamentCard({ tournament }: { tournament: UpcomingTournament }) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-medium">{tournament.name}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-muted-foreground flex items-center gap-2 text-sm">
          <CalendarIcon className="size-4" />
          {formatDateRange(tournament.start_date, tournament.end_date)}
        </div>
      </CardContent>
    </Card>
  );
}

export function UpcomingTournaments() {
  const { data, isLoading } = usePendingActions();

  if (isLoading) {
    return (
      <div className="space-y-3">
        <Skeleton className="h-20 w-full" />
        <Skeleton className="h-20 w-full" />
      </div>
    );
  }

  if (!data?.upcoming_tournaments.length) {
    return null;
  }

  return (
    <div className="space-y-3">
      <h2 className="text-lg font-semibold">Upcoming Tournaments</h2>
      <div className="grid gap-3 sm:grid-cols-2">
        {data.upcoming_tournaments.map((tournament) => (
          <TournamentCard key={tournament.id} tournament={tournament} />
        ))}
      </div>
    </div>
  );
}
