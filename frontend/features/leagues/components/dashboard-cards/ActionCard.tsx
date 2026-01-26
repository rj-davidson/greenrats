"use client";

import { DashboardCard } from "./DashboardCard";
import { Badge } from "@/components/shadcn/badge";
import { Button } from "@/components/shadcn/button";
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/shadcn/empty";
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
  const now = Date.now();

  return (
    [...tournaments]
      .filter((t) => t.pick_window_closes_at && new Date(t.pick_window_closes_at).getTime() > now)
      .sort(
        (a, b) =>
          new Date(a.pick_window_closes_at!).getTime() -
          new Date(b.pick_window_closes_at!).getTime(),
      )[0] ?? null
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
        <Empty className="border-none py-4">
          <EmptyHeader>
            <EmptyTitle>No tournaments scheduled</EmptyTitle>
            <EmptyDescription>Check back soon for upcoming events.</EmptyDescription>
          </EmptyHeader>
        </Empty>
      </DashboardCard>
    );
  }

  const tournament = findRelevantTournament(data.tournaments);

  if (!tournament) {
    return (
      <DashboardCard title="Up Next" icon={<ZapIcon className="size-4" />}>
        <Empty className="border-none py-4">
          <EmptyHeader>
            <EmptyTitle>No upcoming picks</EmptyTitle>
            <EmptyDescription>
              All pick windows are currently closed. Check back soon.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      </DashboardCard>
    );
  }

  const state = getPickWindowState(tournament);

  if (state === "open") {
    const countdown = formatCountdown(tournament.pick_window_closes_at!);

    if (tournament.has_user_pick) {
      return (
        <DashboardCard
          title="Up Next"
          icon={<ZapIcon className="size-4" />}
          className="animate-border-pulse border-2 border-primary"
        >
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <CheckCircle2Icon className="size-5 text-primary" />
              <span className="font-medium">{tournament.name}</span>
            </div>
            <TournamentDetails tournament={tournament} />
            <Link
              href={`/${leagueId}/tournaments/${tournament.id}`}
              className="bg-shimmer block rounded-lg p-3 transition-opacity hover:opacity-80"
            >
              <p className="text-sm text-muted-foreground">Your pick</p>
              <p className="font-semibold">{tournament.golfer_name}</p>
            </Link>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <ClockIcon className="size-4" />
              <span>
                Window closes in <span className="font-medium">{countdown}</span>
              </span>
            </div>
            <Button asChild variant="outline" className="w-full">
              <Link href={`/${leagueId}/tournaments/${tournament.id}`}>View</Link>
            </Button>
          </div>
        </DashboardCard>
      );
    }

    return (
      <DashboardCard
        title="Up Next"
        icon={<ZapIcon className="size-4" />}
        className="animate-border-pulse border-2 border-primary"
      >
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
            <CheckCircle2Icon className="size-5 text-primary" />
            <span className="font-medium">{tournament.name}</span>
          </div>
          <TournamentDetails tournament={tournament} />
          <Link
            href={`/${leagueId}/tournaments/${tournament.id}`}
            className="block rounded-lg bg-muted/50 p-3 transition-opacity hover:opacity-80"
          >
            <p className="text-sm text-muted-foreground">Your pick</p>
            <p className="font-semibold">{tournament.golfer_name}</p>
          </Link>
          <p className="text-sm text-muted-foreground">Pick window closed</p>
          <Button asChild variant="outline" className="w-full">
            <Link href={`/${leagueId}/tournaments/${tournament.id}`}>View</Link>
          </Button>
        </div>
      </DashboardCard>
    );
  }

  return (
    <DashboardCard title="Up Next" icon={<ZapIcon className="size-4" />}>
      <div className="space-y-3">
        <span className="font-medium">{tournament.name}</span>
        <TournamentDetails tournament={tournament} />
        <div className="rounded-lg bg-muted/50 p-3">
          <p className="text-sm text-muted-foreground">Pick window closed</p>
          <p className="font-medium">No selection made</p>
        </div>
      </div>
    </DashboardCard>
  );
}
