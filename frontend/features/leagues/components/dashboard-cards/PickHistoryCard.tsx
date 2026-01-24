"use client";

import { DashboardCard } from "./DashboardCard";
import { Button } from "@/components/shadcn/button";
import { useUserPicks } from "@/features/picks/queries";
import { HistoryIcon } from "lucide-react";
import Link from "next/link";
import { useMemo } from "react";

interface PickHistoryCardProps {
  leagueId: string;
}

function formatEarnings(amount: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(amount);
}

export function PickHistoryCard({ leagueId }: PickHistoryCardProps) {
  const { data, isLoading } = useUserPicks(leagueId);

  const { picks, totalEarnings } = useMemo(() => {
    if (!data?.picks) return { picks: [], totalEarnings: 0 };

    const sortedPicks = [...data.picks].sort((a, b) => {
      const aTime = Date.parse(a.created_at);
      const bTime = Date.parse(b.created_at);
      if (Number.isNaN(aTime) || Number.isNaN(bTime)) {
        return b.created_at.localeCompare(a.created_at);
      }
      return bTime - aTime;
    });

    const total = sortedPicks.reduce((sum, p) => sum + (p.golfer_earnings ?? 0), 0);

    return { picks: sortedPicks.slice(0, 3), totalEarnings: total };
  }, [data]);

  const action = (
    <Button asChild variant="ghost" size="sm">
      <Link href={`/${leagueId}/standings`}>View all</Link>
    </Button>
  );

  if (!picks.length && !isLoading) {
    return (
      <DashboardCard title="Pick History" icon={<HistoryIcon className="size-4" />} action={action}>
        <p className="text-sm text-muted-foreground">You haven&apos;t made any picks yet.</p>
      </DashboardCard>
    );
  }

  return (
    <DashboardCard
      title="Pick History"
      icon={<HistoryIcon className="size-4" />}
      action={action}
      isLoading={isLoading}
    >
      <div className="space-y-3">
        {picks.map((pick) => (
          <div key={pick.id} className="space-y-0.5">
            <div className="text-sm font-medium">{pick.tournament_name}</div>
            <div className="flex items-center justify-between text-sm text-muted-foreground">
              <span>{pick.golfer_name}</span>
              <span className="font-medium text-foreground">
                {pick.golfer_earnings !== undefined ? formatEarnings(pick.golfer_earnings) : "-"}
              </span>
            </div>
          </div>
        ))}
        {picks.length > 0 && (
          <div className="flex items-center justify-between border-t pt-3">
            <span className="font-semibold">Total</span>
            <span className="font-semibold">{formatEarnings(totalEarnings)}</span>
          </div>
        )}
      </div>
    </DashboardCard>
  );
}
