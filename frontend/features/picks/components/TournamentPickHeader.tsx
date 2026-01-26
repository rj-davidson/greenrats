"use client";

import { Badge } from "@/components/shadcn/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/shadcn/card";
import type { PickWindowState } from "@/features/picks/types";
import { formatPickWindowDate, formatCountdown } from "@/features/picks/utils";
import { CalendarIcon, MapPinIcon, TrophyIcon, UserIcon } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

interface TournamentPickHeaderProps {
  tournamentName: string;
  startDate: string;
  endDate: string;
  isLive?: boolean;
  course?: string | null;
  city?: string | null;
  state?: string | null;
  country?: string | null;
  purse?: number | null;
  pickWindowState?: PickWindowState;
  pickWindowOpensAt?: string | null;
  pickWindowClosesAt?: string | null;
  currentPickGolferName?: string;
}

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(amount);
}

function formatDateRange(startDate: string, endDate: string): string {
  const start = new Date(startDate);
  const end = new Date(endDate);
  const startMonth = start.toLocaleDateString("en-US", { month: "short" });
  const endMonth = end.toLocaleDateString("en-US", { month: "short" });
  const startDay = start.getDate();
  const endDay = end.getDate();

  if (startMonth === endMonth) {
    return `${startMonth} ${startDay}-${endDay}`;
  }
  return `${startMonth} ${startDay} - ${endMonth} ${endDay}`;
}

function getCountdownTarget(
  pickWindowState?: PickWindowState,
  pickWindowOpensAt?: string | null,
  pickWindowClosesAt?: string | null,
): string | null {
  if (!pickWindowState || pickWindowState === "closed") return null;
  return pickWindowState === "not_open"
    ? (pickWindowOpensAt ?? null)
    : (pickWindowClosesAt ?? null);
}

export function TournamentPickHeader({
  tournamentName,
  startDate,
  endDate,
  isLive,
  course,
  city,
  state,
  country,
  purse,
  pickWindowState,
  pickWindowOpensAt,
  pickWindowClosesAt,
  currentPickGolferName,
}: TournamentPickHeaderProps) {
  const targetDate = getCountdownTarget(pickWindowState, pickWindowOpensAt, pickWindowClosesAt);

  const initialCountdown = useMemo(
    () => (targetDate ? formatCountdown(targetDate) : ""),
    [targetDate],
  );

  const [countdown, setCountdown] = useState(initialCountdown);

  useEffect(() => {
    setCountdown(initialCountdown);
  }, [initialCountdown]);

  useEffect(() => {
    if (!targetDate) {
      return;
    }

    const interval = setInterval(() => {
      setCountdown(formatCountdown(targetDate));
    }, 60000);
    return () => clearInterval(interval);
  }, [targetDate]);

  const location = [course, city, state, country].filter(Boolean).join(", ");

  const countdownLabel = pickWindowState === "not_open" ? "Opens in" : "Closes in";

  const showPickWindowTimes =
    pickWindowState === "open" && pickWindowOpensAt && pickWindowClosesAt;

  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex flex-wrap items-start justify-between gap-2">
          <CardTitle className="text-xl">{tournamentName}</CardTitle>
          {isLive && (
            <Badge variant="default" className="text-xs">
              Live
            </Badge>
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex flex-wrap gap-x-6 gap-y-2 text-sm text-muted-foreground">
          <div className="flex items-center gap-1.5">
            <CalendarIcon className="size-4" />
            <span>{formatDateRange(startDate, endDate)}</span>
          </div>
          {location && (
            <div className="flex items-center gap-1.5">
              <MapPinIcon className="size-4" />
              <span>{location}</span>
            </div>
          )}
          {purse && (
            <div className="flex items-center gap-1.5">
              <TrophyIcon className="size-4" />
              <span>{formatCurrency(purse)} Purse</span>
            </div>
          )}
        </div>

        {showPickWindowTimes && (
          <div className="flex flex-wrap gap-4 rounded-lg bg-muted/50 p-3 text-sm">
            <div>
              <span className="text-muted-foreground">Opens: </span>
              <span>{formatPickWindowDate(pickWindowOpensAt)}</span>
            </div>
            <div>
              <span className="text-muted-foreground">Closes: </span>
              <span>{formatPickWindowDate(pickWindowClosesAt)}</span>
            </div>
            {countdown && (
              <div className="font-medium">
                <span className="text-muted-foreground">{countdownLabel}: </span>
                <span className="text-primary">{countdown}</span>
              </div>
            )}
          </div>
        )}

        {currentPickGolferName && (
          <div className="flex items-center gap-3 rounded-lg border bg-primary/5 p-3">
            <UserIcon className="size-5 text-primary" />
            <div>
              <div className="text-sm font-medium">Your Current Pick</div>
              <div className="text-sm text-muted-foreground">{currentPickGolferName}</div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
