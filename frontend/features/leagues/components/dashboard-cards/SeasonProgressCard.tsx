"use client";

import { DashboardCard } from "./DashboardCard";
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/shadcn/empty";
import { Progress } from "@/components/shadcn/progress";
import { useLeagueTournaments } from "@/features/leagues/queries";
import type { LeagueTournament } from "@/features/leagues/types";
import { CalendarDaysIcon } from "lucide-react";
import { useMemo } from "react";

interface SeasonProgressCardProps {
  leagueId: string;
}

interface SeasonStats {
  completedCount: number;
  total: number;
  progress: number;
  currentOrNext: LeagueTournament | undefined;
}

export function SeasonProgressCard({ leagueId }: SeasonProgressCardProps) {
  const { data, isLoading } = useLeagueTournaments(leagueId);

  const stats = useMemo((): SeasonStats | null => {
    if (!data?.tournaments) return null;

    const now = new Date();
    const started = data.tournaments.filter((t) => new Date(t.start_date) < now);
    const active = data.tournaments.find((t) => t.status === "active");
    const upcoming = data.tournaments
      .filter((t) => t.status === "upcoming")
      .sort((a, b) => new Date(a.start_date).getTime() - new Date(b.start_date).getTime());

    const total = data.tournaments.length;
    const completedCount = started.length;
    const progress = total > 0 ? (completedCount / total) * 100 : 0;

    return {
      completedCount,
      total,
      progress,
      currentOrNext: active ?? upcoming[0],
    };
  }, [data]);

  return (
    <DashboardCard
      title="Season Progress"
      icon={<CalendarDaysIcon className="size-4" />}
      isLoading={isLoading}
    >
      {stats ? (
        <div className="space-y-4">
          <div>
            <div className="mb-2 flex items-baseline justify-between">
              <span className="text-2xl font-bold">{stats.completedCount}</span>
              <span className="text-sm text-muted-foreground">of {stats.total} tournaments</span>
            </div>
            <Progress value={stats.progress} />
          </div>
          {stats.currentOrNext && (
            <div className="rounded-lg bg-muted/50 p-3">
              <p className="text-xs text-muted-foreground">
                {stats.currentOrNext.status === "active" ? "Current" : "Next"}
              </p>
              <p className="font-medium">{stats.currentOrNext.name}</p>
            </div>
          )}
        </div>
      ) : (
        <Empty className="border-none py-4">
          <EmptyHeader>
            <EmptyTitle>No schedule yet</EmptyTitle>
            <EmptyDescription>
              The tournament schedule will appear once it&apos;s published.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      )}
    </DashboardCard>
  );
}
