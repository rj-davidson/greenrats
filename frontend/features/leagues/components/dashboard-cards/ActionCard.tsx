"use client";

import { DashboardCard } from "./DashboardCard";
import { Badge } from "@/components/shadcn/badge";
import { Button } from "@/components/shadcn/button";
import { useLeagueTournaments } from "@/features/leagues/queries";
import type { LeagueTournament } from "@/features/leagues/types";
import { formatCountdown, getPickWindowState } from "@/features/picks/utils";
import {
  CalendarIcon,
  CheckCircle2Icon,
  ClockIcon,
  FlagIcon,
  MapPinIcon,
  ZapIcon,
} from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";

interface ActionCardProps {
  leagueId: string;
}

function findRelevantTournament(tournaments: LeagueTournament[]): LeagueTournament | null {
  const sorted = [...tournaments].sort((a, b) => {
    const aTime = a.pick_window_closes_at
      ? new Date(a.pick_window_closes_at).getTime()
      : new Date(a.start_date).getTime();
    const bTime = b.pick_window_closes_at
      ? new Date(b.pick_window_closes_at).getTime()
      : new Date(b.start_date).getTime();
    return aTime - bTime;
  });

  const active = sorted.find(
    (t) => t.status === "active" && t.pick_window_opens_at && t.pick_window_closes_at,
  );
  if (active) {
    const state = getPickWindowState(active);
    if (state === "open") return active;
  }

  return (
    sorted.find(
      (t) => t.status === "upcoming" && t.pick_window_opens_at && t.pick_window_closes_at,
    ) ?? null
  );
}

function formatLocalDateTime(isoString: string): string {
  const date = new Date(isoString);
  return date.toLocaleDateString("en-US", {
    weekday: "short",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
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

function formatLocation(city?: string, state?: string, country?: string): string | null {
  if (city && state) return `${city}, ${state}`;
  if (city && country) return `${city}, ${country}`;
  if (city) return city;
  if (state) return state;
  if (country) return country;
  return null;
}

function TournamentDetails({ tournament }: { tournament: LeagueTournament }) {
  const location = formatLocation(tournament.city, tournament.state, tournament.country);
  const dateRange = formatDateRange(tournament.start_date, tournament.end_date);

  return (
    <div className="space-y-1 text-sm text-muted-foreground">
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
        <span>{dateRange}</span>
      </div>
    </div>
  );
}

export function ActionCard({ leagueId }: ActionCardProps) {
  const { data, isLoading } = useLeagueTournaments(leagueId);
  const [, setTick] = useState(0);

  useEffect(() => {
    const interval = setInterval(() => setTick((t) => t + 1), 60000);
    return () => clearInterval(interval);
  }, []);

  if (isLoading) {
    return <DashboardCard title="Up Next" icon={<ZapIcon className="size-4" />} isLoading />;
  }

  if (!data?.tournaments.length) {
    return (
      <DashboardCard title="Up Next" icon={<ZapIcon className="size-4" />}>
        <p className="text-sm text-muted-foreground">No upcoming tournaments</p>
      </DashboardCard>
    );
  }

  const tournament = findRelevantTournament(data.tournaments);

  if (!tournament) {
    return (
      <DashboardCard title="Up Next" icon={<ZapIcon className="size-4" />}>
        <p className="text-sm text-muted-foreground">No upcoming tournaments</p>
      </DashboardCard>
    );
  }

  const state = getPickWindowState(tournament);

  if (state === "open") {
    const countdown = formatCountdown(tournament.pick_window_closes_at!);

    if (tournament.has_user_pick) {
      return (
        <DashboardCard title="Up Next" icon={<ZapIcon className="size-4" />}>
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <CheckCircle2Icon className="size-5 text-green-600" />
              <span className="font-medium">{tournament.name}</span>
            </div>
            <TournamentDetails tournament={tournament} />
            <div className="rounded-lg bg-muted/50 p-3">
              <p className="text-sm text-muted-foreground">Your pick</p>
              <p className="font-semibold">{tournament.golfer_name}</p>
            </div>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <ClockIcon className="size-4" />
                <span>
                  Window closes in <span className="font-medium">{countdown}</span>
                </span>
              </div>
              <Button asChild size="sm" variant="outline">
                <Link href={`/${leagueId}/tournaments/${tournament.id}`}>View Field</Link>
              </Button>
            </div>
          </div>
        </DashboardCard>
      );
    }

    return (
      <DashboardCard title="Up Next" icon={<ZapIcon className="size-4" />}>
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <span className="font-medium">{tournament.name}</span>
            <Badge variant="default" className="text-xs">
              Open
            </Badge>
          </div>
          <TournamentDetails tournament={tournament} />
          <div className="flex items-center gap-2 text-sm">
            <ClockIcon className="size-4 text-muted-foreground" />
            <span>
              Closes in <span className="font-semibold">{countdown}</span>
            </span>
          </div>
          <Button asChild className="w-full">
            <Link href={`/${leagueId}/tournaments/${tournament.id}`}>Make Your Pick</Link>
          </Button>
        </div>
      </DashboardCard>
    );
  }

  if (state === "not_open") {
    const countdown = formatCountdown(tournament.pick_window_opens_at!);
    const localDateTime = formatLocalDateTime(tournament.pick_window_opens_at!);

    return (
      <DashboardCard title="Up Next" icon={<ZapIcon className="size-4" />}>
        <div className="space-y-3">
          <span className="font-medium">{tournament.name}</span>
          <TournamentDetails tournament={tournament} />
          <div className="rounded-lg bg-muted/50 p-3">
            <p className="text-sm text-muted-foreground">Pick window opens</p>
            <p className="font-semibold">{countdown}</p>
            <p className="mt-1 text-xs text-muted-foreground">{localDateTime}</p>
          </div>
        </div>
      </DashboardCard>
    );
  }

  if (tournament.has_user_pick) {
    return (
      <DashboardCard title="Up Next" icon={<ZapIcon className="size-4" />}>
        <div className="space-y-3">
          <div className="flex items-center gap-2">
            <CheckCircle2Icon className="size-5 text-green-600" />
            <span className="font-medium">{tournament.name}</span>
          </div>
          <TournamentDetails tournament={tournament} />
          <div className="rounded-lg bg-muted/50 p-3">
            <p className="text-sm text-muted-foreground">Your pick</p>
            <p className="font-semibold">{tournament.golfer_name}</p>
          </div>
          <p className="text-sm text-muted-foreground">Pick window closed</p>
        </div>
      </DashboardCard>
    );
  }

  return (
    <DashboardCard title="Up Next" icon={<ZapIcon className="size-4" />}>
      <div className="space-y-3">
        <span className="font-medium">{tournament.name}</span>
        <TournamentDetails tournament={tournament} />
        <p className="text-sm text-muted-foreground">No pick made for this tournament</p>
      </div>
    </DashboardCard>
  );
}
