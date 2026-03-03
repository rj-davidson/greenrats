"use client";

import { DashboardCard } from "./DashboardCard";
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

function findRelevantTournaments(tournaments: LeagueTournament[]): LeagueTournament[] {
  const now = Date.now();

  const withOpenWindow = [...tournaments]
    .filter(
      (t) =>
        t.pick_window_opens_at &&
        t.pick_window_closes_at &&
        new Date(t.pick_window_opens_at).getTime() <= now &&
        new Date(t.pick_window_closes_at).getTime() > now,
    )
    .sort(
      (a, b) =>
        new Date(a.pick_window_closes_at!).getTime() - new Date(b.pick_window_closes_at!).getTime(),
    );

  if (withOpenWindow.length > 0) return withOpenWindow;

  const nextUpcoming = [...tournaments]
    .filter((t) => t.pick_window_closes_at && new Date(t.pick_window_closes_at).getTime() > now)
    .sort(
      (a, b) =>
        new Date(a.pick_window_closes_at!).getTime() - new Date(b.pick_window_closes_at!).getTime(),
    )[0];

  return nextUpcoming ? [nextUpcoming] : [];
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

function ActionCardContent({
  leagueId,
  tournament,
}: {
  leagueId: string;
  tournament: LeagueTournament;
}) {
  const state = getPickWindowState(tournament);
  const isOpen = state === "open";
  const isNotOpen = state === "not_open";
  const notInField = tournament.has_user_pick && tournament.golfer_in_field === false;

  const pickSlotBg = notInField
    ? "rounded-lg border border-destructive/20 bg-destructive/10 p-3"
    : isOpen && tournament.has_user_pick
      ? "bg-shimmer rounded-lg p-3"
      : "rounded-lg bg-muted/50 p-3";

  const pickSlot = (() => {
    if (tournament.has_user_pick) {
      return (
        <Link
          href={`/${leagueId}/tournaments/${tournament.id}`}
          className={`block transition-opacity hover:opacity-80 ${pickSlotBg}`}
        >
          <p className="text-sm text-muted-foreground">Your pick</p>
          <p className="font-semibold">
            {tournament.golfer_name}
            {notInField && <span className="text-destructive"> (Not in field)</span>}
          </p>
        </Link>
      );
    }

    if (isNotOpen) {
      return (
        <div className={pickSlotBg}>
          <p className="text-sm text-muted-foreground">Pick window opens</p>
          <p className="font-semibold">{formatCountdown(tournament.pick_window_opens_at!)}</p>
          <p className="mt-1 text-xs text-muted-foreground">
            {formatLocalDateTime(tournament.pick_window_opens_at!)}
          </p>
        </div>
      );
    }

    if (!isOpen) {
      return (
        <div className={pickSlotBg}>
          <p className="text-sm text-muted-foreground">No selection made</p>
        </div>
      );
    }

    return null;
  })();

  const timerRow = isOpen ? (
    <div className="flex items-center gap-2 text-sm text-muted-foreground">
      <ClockIcon className="size-4" />
      <span>
        Closes in{" "}
        <span className="font-medium">{formatCountdown(tournament.pick_window_closes_at!)}</span>
      </span>
    </div>
  ) : null;

  const actionButton = (() => {
    if (isOpen && !tournament.has_user_pick) {
      return (
        <Button asChild className="w-full">
          <Link href={`/${leagueId}/tournaments/${tournament.id}`}>Make Your Pick</Link>
        </Button>
      );
    }
    if (isOpen && tournament.has_user_pick) {
      return (
        <Button asChild variant="outline" className="w-full">
          <Link href={`/${leagueId}/tournaments/${tournament.id}`}>Change Pick</Link>
        </Button>
      );
    }
    if (!isOpen && !isNotOpen && tournament.has_user_pick) {
      return (
        <Button asChild variant="outline" className="w-full">
          <Link href={`/${leagueId}/tournaments/${tournament.id}`}>View</Link>
        </Button>
      );
    }
    return null;
  })();

  return (
    <DashboardCard
      title="Up Next"
      icon={<ZapIcon className="size-4" />}
      className={isOpen ? "animate-border-pulse border-2 border-primary" : undefined}
    >
      <div className="space-y-3">
        <div className="flex items-center gap-2">
          {tournament.has_user_pick && <CheckCircle2Icon className="size-5 text-primary" />}
          <span className="font-medium">{tournament.name}</span>
        </div>
        <TournamentDetails tournament={tournament} />
        {pickSlot}
        {timerRow}
        {actionButton}
      </div>
    </DashboardCard>
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

  const tournaments = findRelevantTournaments(data.tournaments);

  if (tournaments.length === 0) {
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

  return (
    <>
      {tournaments.map((tournament) => (
        <ActionCardContent key={tournament.id} leagueId={leagueId} tournament={tournament} />
      ))}
    </>
  );
}
