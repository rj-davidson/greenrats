"use client";

import { Button } from "@/components/shadcn/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/shadcn/card";
import { Skeleton } from "@/components/shadcn/skeleton";
import { usePendingActions } from "@/features/users/queries";
import type { PendingPickAction } from "@/features/users/types";
import { AlertCircleIcon, ClockIcon } from "lucide-react";
import Link from "next/link";

function formatDeadline(deadline: string): string {
  const date = new Date(deadline);
  const now = new Date();
  const hoursUntil = (date.getTime() - now.getTime()) / (1000 * 60 * 60);

  if (hoursUntil < 24) {
    return `${Math.round(hoursUntil)} hours`;
  }
  const daysUntil = Math.round(hoursUntil / 24);
  return `${daysUntil} day${daysUntil !== 1 ? "s" : ""}`;
}

function PendingPickCard({ action }: { action: PendingPickAction }) {
  return (
    <Card className="border-amber-200 bg-amber-50 dark:border-amber-900 dark:bg-amber-950/30">
      <CardHeader className="pb-2">
        <CardTitle className="flex items-center justify-between text-base">
          <span className="flex items-center gap-2">
            <AlertCircleIcon className="size-4 text-amber-600" />
            Pick needed for {action.tournament_name}
          </span>
          <span className="flex items-center gap-1 text-sm font-normal text-muted-foreground">
            <ClockIcon className="size-3" />
            {formatDeadline(action.pick_deadline)}
          </span>
        </CardTitle>
      </CardHeader>
      <CardContent className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">
          Make your pick for {action.league_name}
        </span>
        <Button asChild size="sm">
          <Link href={`/${action.league_id}/tournaments/${action.tournament_id}`}>Make Pick</Link>
        </Button>
      </CardContent>
    </Card>
  );
}

export function PendingActions() {
  const { data, isLoading } = usePendingActions();

  if (isLoading) {
    return (
      <div className="space-y-3">
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
      </div>
    );
  }

  if (!data?.pending_picks.length) {
    return null;
  }

  return (
    <div className="space-y-3">
      <h2 className="text-lg font-semibold">Pending Actions</h2>
      {data.pending_picks.map((action) => (
        <PendingPickCard key={`${action.league_id}-${action.tournament_id}`} action={action} />
      ))}
    </div>
  );
}
