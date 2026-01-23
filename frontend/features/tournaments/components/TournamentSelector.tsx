"use client";

import { Badge } from "@/components/shadcn/badge";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/shadcn/select";
import { Skeleton } from "@/components/shadcn/skeleton";
import { useLeagueTournaments } from "@/features/leagues/queries";
import type { LeagueTournament } from "@/features/leagues/types";
import { format } from "date-fns";
import { useRouter } from "next/navigation";
import { useMemo } from "react";

interface TournamentSelectorProps {
  leagueId: string;
  currentTournamentId?: string;
}

function TournamentStatusBadge({ status }: { status: string }) {
  if (status === "active") {
    return (
      <Badge variant="default" className="ml-auto text-xs">
        Live
      </Badge>
    );
  }
  if (status === "upcoming") {
    return (
      <Badge variant="secondary" className="ml-auto text-xs">
        Upcoming
      </Badge>
    );
  }
  return null;
}

function groupTournaments(tournaments: LeagueTournament[]) {
  const active: LeagueTournament[] = [];
  const upcoming: LeagueTournament[] = [];
  const completed: LeagueTournament[] = [];

  for (const t of tournaments) {
    if (t.status === "active") {
      active.push(t);
    } else if (t.status === "upcoming") {
      upcoming.push(t);
    } else {
      completed.push(t);
    }
  }

  return { active, upcoming, completed };
}

export function TournamentSelector({ leagueId, currentTournamentId }: TournamentSelectorProps) {
  const { data, isLoading } = useLeagueTournaments(leagueId);
  const router = useRouter();

  const { active, upcoming, completed } = useMemo(() => {
    if (!data?.tournaments) return { active: [], upcoming: [], completed: [] };
    return groupTournaments(data.tournaments);
  }, [data]);

  const currentTournament = data?.tournaments.find((t) => t.id === currentTournamentId);

  if (isLoading) {
    return <Skeleton className="h-9 w-64" />;
  }

  const handleSelect = (tournamentId: string) => {
    router.push(`/${leagueId}/tournaments/${tournamentId}`);
  };

  return (
    <Select value={currentTournamentId} onValueChange={handleSelect}>
      <SelectTrigger className="w-64">
        <SelectValue placeholder="Select tournament">
          {currentTournament && (
            <div className="flex items-center gap-2">
              <span className="truncate">{currentTournament.name}</span>
              <TournamentStatusBadge status={currentTournament.status} />
            </div>
          )}
        </SelectValue>
      </SelectTrigger>
      <SelectContent>
        {active.length > 0 && (
          <SelectGroup>
            <SelectLabel>Active</SelectLabel>
            {active.map((tournament) => (
              <SelectItem key={tournament.id} value={tournament.id}>
                <div className="flex w-full items-center justify-between gap-4">
                  <span>{tournament.name}</span>
                  <span className="text-xs text-muted-foreground">
                    {format(new Date(tournament.start_date), "MMM d")} -{" "}
                    {format(new Date(tournament.end_date), "MMM d")}
                  </span>
                </div>
              </SelectItem>
            ))}
          </SelectGroup>
        )}
        {upcoming.length > 0 && (
          <SelectGroup>
            <SelectLabel>Upcoming</SelectLabel>
            {upcoming.map((tournament) => (
              <SelectItem key={tournament.id} value={tournament.id}>
                <div className="flex w-full items-center justify-between gap-4">
                  <span>{tournament.name}</span>
                  <span className="text-xs text-muted-foreground">
                    {format(new Date(tournament.start_date), "MMM d")} -{" "}
                    {format(new Date(tournament.end_date), "MMM d")}
                  </span>
                </div>
              </SelectItem>
            ))}
          </SelectGroup>
        )}
        {completed.length > 0 && (
          <SelectGroup>
            <SelectLabel>Completed</SelectLabel>
            {completed.map((tournament) => (
              <SelectItem key={tournament.id} value={tournament.id}>
                <div className="flex w-full items-center justify-between gap-4">
                  <span>{tournament.name}</span>
                  <span className="text-xs text-muted-foreground">
                    {format(new Date(tournament.start_date), "MMM d")} -{" "}
                    {format(new Date(tournament.end_date), "MMM d")}
                  </span>
                </div>
              </SelectItem>
            ))}
          </SelectGroup>
        )}
      </SelectContent>
    </Select>
  );
}
