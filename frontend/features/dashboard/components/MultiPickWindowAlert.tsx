"use client";

import { Alert, AlertDescription, AlertTitle } from "@/components/shadcn/alert";
import { getPickWindowState } from "@/features/picks/utils";
import { useTournaments } from "@/features/tournaments/queries";
import { RatIcon } from "lucide-react";

export function MultiPickWindowAlert() {
  const { data } = useTournaments({ status: "upcoming" });

  const openCount = data?.tournaments.filter((t) => getPickWindowState(t) === "open").length ?? 0;

  if (openCount < 2) return null;

  return (
    <Alert className="mb-6 border-secondary-foreground/20 bg-secondary text-secondary-foreground">
      <RatIcon className="size-4" />
      <AlertTitle>{openCount} pick windows open this week</AlertTitle>
      <AlertDescription className="text-secondary-foreground">
        Make sure to pick a golfer for each tournament before the windows close.
      </AlertDescription>
    </Alert>
  );
}
