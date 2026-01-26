"use client";

import { DashboardCard } from "./DashboardCard";
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/shadcn/empty";
import { useLeagueLeaderboard } from "@/features/leaderboards/queries";
import { useCurrentUser } from "@/features/users/queries";
import { TrophyIcon } from "lucide-react";

interface YourStatsCardProps {
  leagueId: string;
}

function formatEarnings(amount: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(amount);
}

export function YourStatsCard({ leagueId }: YourStatsCardProps) {
  const { data: currentUser } = useCurrentUser();
  const { data: leaderboardData, isLoading } = useLeagueLeaderboard(leagueId);

  const userEntry = leaderboardData?.entries.find((e) => e.user_id === currentUser?.id);
  const leaderEntry = leaderboardData?.entries.find((e) => e.rank === 1);
  const totalMembers = leaderboardData?.entries.length ?? 0;

  const gapFromFirst =
    userEntry && leaderEntry && userEntry.rank > 1
      ? leaderEntry.earnings - userEntry.earnings
      : null;

  return (
    <DashboardCard
      title="Your Stats"
      icon={<TrophyIcon className="size-4" />}
      isLoading={isLoading}
    >
      {userEntry ? (
        <div className="space-y-4">
          <div className="flex items-baseline gap-2">
            <span className="text-4xl font-bold">{userEntry.rank_display}</span>
            <span className="text-sm text-muted-foreground">of {totalMembers} members</span>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Total Earnings</p>
            <p className="text-2xl font-semibold">{formatEarnings(userEntry.earnings)}</p>
          </div>
          {gapFromFirst !== null ? (
            <p className="text-sm text-muted-foreground">
              {formatEarnings(gapFromFirst)} behind leader
            </p>
          ) : userEntry.rank === 1 ? (
            <p className="text-sm font-medium text-primary">You&apos;re in the lead!</p>
          ) : null}
        </div>
      ) : (
        <Empty className="border-none py-4">
          <EmptyHeader>
            <EmptyTitle>No stats yet</EmptyTitle>
            <EmptyDescription>
              Your rankings and earnings will appear after your first pick.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      )}
    </DashboardCard>
  );
}
