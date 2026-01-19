"use client";

import { useLeagueLeaderboard } from "@/features/leaderboards/queries";
import { Skeleton } from "@/components/shadcn/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/shadcn/table";
import { useCurrentUser } from "@/features/users/queries";
import { cn } from "@/lib/utils";
import { TrophyIcon } from "lucide-react";

interface LeagueLeaderboardProps {
  leagueId: string;
}

function formatEarnings(amount: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(amount);
}

function getRankDisplay(rank: number) {
  if (rank === 1) {
    return (
      <span className="flex items-center gap-1 text-amber-500">
        <TrophyIcon className="size-4" />1
      </span>
    );
  }
  if (rank === 2) {
    return <span className="text-slate-400">{rank}</span>;
  }
  if (rank === 3) {
    return <span className="text-amber-700">{rank}</span>;
  }
  return rank;
}

export function LeagueLeaderboard({ leagueId }: LeagueLeaderboardProps) {
  const { data, isLoading, error } = useLeagueLeaderboard(leagueId);
  const { data: currentUser } = useCurrentUser();

  if (isLoading) {
    return (
      <div className="space-y-2">
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-10 w-full" />
      </div>
    );
  }

  if (error) {
    return <div className="text-destructive">Failed to load leaderboard</div>;
  }

  if (!data?.entries.length) {
    return (
      <div className="py-8 text-center text-muted-foreground">
        No picks have been made yet. The leaderboard will appear once members start making picks.
      </div>
    );
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-16">Rank</TableHead>
          <TableHead>Player</TableHead>
          <TableHead className="text-right">Picks</TableHead>
          <TableHead className="text-right">Earnings</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {data.entries.map((entry) => {
          const isCurrentUser = currentUser?.id === entry.user_id;
          return (
            <TableRow
              key={entry.user_id}
              className={cn(isCurrentUser && "bg-primary/5 font-medium")}
            >
              <TableCell>{getRankDisplay(entry.rank)}</TableCell>
              <TableCell>
                {entry.display_name}
                {isCurrentUser && <span className="ml-2 text-sm text-muted-foreground">(you)</span>}
              </TableCell>
              <TableCell className="text-right">{entry.pick_count}</TableCell>
              <TableCell className="text-right">{formatEarnings(entry.earnings)}</TableCell>
            </TableRow>
          );
        })}
      </TableBody>
    </Table>
  );
}
