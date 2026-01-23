"use client";

import { Badge } from "@/components/shadcn/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/shadcn/card";
import type { GetPickFieldResponse, PickWindowState } from "@/features/picks/types";
import { formatPickWindowDate, formatCountdown } from "@/features/picks/utils";
import { CalendarIcon, ClockIcon, LockIcon, MapPinIcon, TrophyIcon, UnlockIcon, UserIcon } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

interface TournamentPickHeaderProps {
  data: GetPickFieldResponse;
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

function getPickWindowBadge(state: PickWindowState) {
  switch (state) {
    case "open":
      return (
        <Badge variant="default" className="gap-1">
          <UnlockIcon className="size-3" />
          Open
        </Badge>
      );
    case "not_open":
      return (
        <Badge variant="secondary" className="gap-1">
          <ClockIcon className="size-3" />
          Not Yet Open
        </Badge>
      );
    case "closed":
      return (
        <Badge variant="secondary" className="gap-1">
          <LockIcon className="size-3" />
          Closed
        </Badge>
      );
  }
}

function getCountdownTarget(data: GetPickFieldResponse): string | null {
  if (data.pick_window_state === "closed") return null;
  return data.pick_window_state === "not_open"
    ? data.pick_window_opens_at ?? null
    : data.pick_window_closes_at ?? null;
}

export function TournamentPickHeader({ data, currentPickGolferName }: TournamentPickHeaderProps) {
  const targetDate = getCountdownTarget(data);

  const initialCountdown = useMemo(
    () => (targetDate ? formatCountdown(targetDate) : ""),
    [targetDate]
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

  const location = [data.course, data.city, data.state, data.country]
    .filter(Boolean)
    .join(", ");

  const countdownLabel = data.pick_window_state === "not_open" ? "Opens in" : "Closes in";

  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex flex-wrap items-start justify-between gap-2">
          <CardTitle className="text-xl">{data.tournament_name}</CardTitle>
          {getPickWindowBadge(data.pick_window_state)}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex flex-wrap gap-x-6 gap-y-2 text-sm text-muted-foreground">
          <div className="flex items-center gap-1.5">
            <CalendarIcon className="size-4" />
            <span>{formatDateRange(data.start_date, data.end_date)}</span>
          </div>
          {location && (
            <div className="flex items-center gap-1.5">
              <MapPinIcon className="size-4" />
              <span>{location}</span>
            </div>
          )}
          {data.purse && (
            <div className="flex items-center gap-1.5">
              <TrophyIcon className="size-4" />
              <span>{formatCurrency(data.purse)} Purse</span>
            </div>
          )}
        </div>

        {data.pick_window_state !== "closed" && data.pick_window_opens_at && data.pick_window_closes_at && (
          <div className="flex flex-wrap gap-4 rounded-lg bg-muted/50 p-3 text-sm">
            <div>
              <span className="text-muted-foreground">Opens: </span>
              <span>{formatPickWindowDate(data.pick_window_opens_at)}</span>
            </div>
            <div>
              <span className="text-muted-foreground">Closes: </span>
              <span>{formatPickWindowDate(data.pick_window_closes_at)}</span>
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
