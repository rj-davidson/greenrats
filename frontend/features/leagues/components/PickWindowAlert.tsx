"use client";

import { Alert, AlertDescription, AlertTitle } from "@/components/shadcn/alert";
import { Button } from "@/components/shadcn/button";
import { useLeagueTournaments } from "@/features/leagues/queries";
import type { LeagueTournament } from "@/features/leagues/types";
import { formatCountdown, getPickWindowState } from "@/features/picks/utils";
import { CalendarIcon, ClockIcon } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";

interface PickWindowAlertProps {
  leagueId: string;
}

function formatLocalDateTime(isoString: string): string {
  const date = new Date(isoString);
  return date.toLocaleDateString("en-US", {
    weekday: "long",
    month: "long",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
    timeZoneName: "short",
  });
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

function OpenWindowAlert({
  tournament,
  leagueId,
}: {
  tournament: LeagueTournament;
  leagueId: string;
}) {
  const countdown = formatCountdown(tournament.pick_window_closes_at!);

  if (tournament.has_user_pick) {
    return (
      <Alert className="mb-6 border-primary">
        <ClockIcon className="size-4" />
        <AlertTitle>{tournament.name}</AlertTitle>
        <AlertDescription>
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="space-y-1">
              <p>
                Your pick: <span className="font-medium">{tournament.golfer_name}</span>
              </p>
              <p className="text-muted-foreground">
                Window closes in <span className="font-medium">{countdown}</span>
              </p>
            </div>
            <Button asChild size="sm" variant="outline">
              <Link href={`/${leagueId}/tournaments/${tournament.id}`}>View Pick</Link>
            </Button>
          </div>
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <Alert className="animate-shimmer mb-6 border-primary">
      <ClockIcon className="size-4" />
      <AlertTitle>Pick window open for {tournament.name}</AlertTitle>
      <AlertDescription>
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <p>
            Closes in <span className="font-medium">{countdown}</span>
          </p>
          <Button asChild size="sm">
            <Link href={`/${leagueId}/tournaments/${tournament.id}`}>Make Your Pick</Link>
          </Button>
        </div>
      </AlertDescription>
    </Alert>
  );
}

function NotOpenWindowAlert({ tournament }: { tournament: LeagueTournament }) {
  const countdown = formatCountdown(tournament.pick_window_opens_at!);
  const localDateTime = formatLocalDateTime(tournament.pick_window_opens_at!);

  return (
    <Alert className="mb-6">
      <CalendarIcon className="size-4" />
      <AlertTitle>Pick window opens for {tournament.name}</AlertTitle>
      <AlertDescription>
        <div className="space-y-1">
          <p>
            Opens in <span className="font-medium">{countdown}</span>
          </p>
          <p className="text-muted-foreground">{localDateTime}</p>
        </div>
      </AlertDescription>
    </Alert>
  );
}

export function PickWindowAlert({ leagueId }: PickWindowAlertProps) {
  const { data, isLoading } = useLeagueTournaments(leagueId);
  const [_tick, setTick] = useState(0);

  useEffect(() => {
    const interval = setInterval(() => setTick((t) => t + 1), 60000);
    return () => clearInterval(interval);
  }, []);

  if (isLoading || !data?.tournaments) {
    return null;
  }

  const tournament = findRelevantTournament(data.tournaments);
  if (!tournament) {
    return null;
  }

  const state = getPickWindowState(tournament);

  if (state === "open") {
    return <OpenWindowAlert tournament={tournament} leagueId={leagueId} />;
  }

  if (state === "not_open") {
    return <NotOpenWindowAlert tournament={tournament} />;
  }

  return null;
}
