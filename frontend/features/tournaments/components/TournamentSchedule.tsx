"use client";

import { Badge } from "@/components/shadcn/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/shadcn/card";
import { Skeleton } from "@/components/shadcn/skeleton";
import { useTournaments } from "@/features/tournaments/queries";
import type { Tournament, TournamentStatus } from "@/features/tournaments/types";
import { CalendarIcon, MapPinIcon } from "lucide-react";

function formatDateRange(startDate: string, endDate: string): string {
  const start = new Date(startDate);
  const end = new Date(endDate);
  const options: Intl.DateTimeFormatOptions = { month: "short", day: "numeric" };

  if (start.getMonth() === end.getMonth()) {
    return `${start.toLocaleDateString("en-US", { month: "short", day: "numeric" })} - ${end.getDate()}, ${end.getFullYear()}`;
  }

  return `${start.toLocaleDateString("en-US", options)} - ${end.toLocaleDateString("en-US", options)}, ${end.getFullYear()}`;
}

function getStatusBadgeVariant(status: TournamentStatus): "default" | "secondary" | "outline" {
  switch (status) {
    case "active":
      return "default";
    case "completed":
      return "secondary";
    case "upcoming":
    default:
      return "outline";
  }
}

function getStatusLabel(status: TournamentStatus): string {
  switch (status) {
    case "active":
      return "Live";
    case "completed":
      return "Completed";
    case "upcoming":
    default:
      return "Upcoming";
  }
}

function TournamentCard({ tournament }: { tournament: Tournament }) {
  const venue = tournament.venue || tournament.course;

  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between gap-2">
          <CardTitle className="text-lg">{tournament.name}</CardTitle>
          <Badge variant={getStatusBadgeVariant(tournament.status)}>
            {getStatusLabel(tournament.status)}
          </Badge>
        </div>
        <CardDescription className="flex items-center gap-1">
          <CalendarIcon className="size-3" />
          {formatDateRange(tournament.start_date, tournament.end_date)}
        </CardDescription>
      </CardHeader>
      <CardContent>
        {venue && (
          <p className="text-muted-foreground flex items-center gap-1 text-sm">
            <MapPinIcon className="size-3" />
            {venue}
          </p>
        )}
        {tournament.purse && (
          <p className="text-muted-foreground mt-1 text-sm">
            Purse: ${tournament.purse.toLocaleString()}
          </p>
        )}
      </CardContent>
    </Card>
  );
}

function TournamentSkeleton() {
  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between gap-2">
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-5 w-20" />
        </div>
        <Skeleton className="mt-2 h-4 w-32" />
      </CardHeader>
      <CardContent>
        <Skeleton className="h-4 w-40" />
      </CardContent>
    </Card>
  );
}

export function TournamentSchedule() {
  const { data, isLoading, error } = useTournaments();

  if (isLoading) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 6 }).map((_, i) => (
          <TournamentSkeleton key={i} />
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground">Failed to load tournaments. Please try again.</p>
        </CardContent>
      </Card>
    );
  }

  if (!data || !data.tournaments.length) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground">No tournaments scheduled.</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {data.tournaments.map((tournament) => (
        <TournamentCard key={tournament.id} tournament={tournament} />
      ))}
    </div>
  );
}
