"use client";

import { DashboardCard } from "./DashboardCard";
import { Badge } from "@/components/shadcn/badge";
import { Button } from "@/components/shadcn/button";
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/shadcn/empty";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/shadcn/table";
import { useLeagueTournaments } from "@/features/leagues/queries";
import { useLeaguePicks } from "@/features/picks/queries";
import { formatScoreToPar } from "@/features/tournaments/components/leaderboard-utils";
import { useLeaderboard } from "@/features/tournaments/queries";
import { useCurrentUser } from "@/features/users/queries";
import { cn } from "@/lib/utils";
import { TrophyIcon } from "lucide-react";
import Link from "next/link";
import { useMemo } from "react";

interface RecentTournamentResultsCardProps {
  leagueId: string;
}

function formatEarnings(amount: number): string {
  if (amount >= 1_000_000) {
    return `$${(amount / 1_000_000).toFixed(2)}M`;
  }
  if (amount >= 1_000) {
    return `$${(amount / 1_000).toFixed(0)}K`;
  }
  return `$${amount.toLocaleString()}`;
}

function pickDate(date?: string) {
  return Date.parse(date ?? "");
}

export function RecentTournamentResultsCard({ leagueId }: RecentTournamentResultsCardProps) {
  const { data: tournamentsData, isLoading: tournamentsLoading } = useLeagueTournaments(leagueId);
  const { data: currentUser } = useCurrentUser();

  const recentCompleted = useMemo(() => {
    if (!tournamentsData?.tournaments) return null;
    const completed = tournamentsData.tournaments.filter((t) => t.status === "completed");
    if (completed.length === 0) return null;

    return [...completed].sort((a, b) => {
      const aTime = pickDate(a.end_date || a.start_date);
      const bTime = pickDate(b.end_date || b.start_date);
      if (Number.isNaN(aTime) || Number.isNaN(bTime)) {
        return (b.end_date || b.start_date).localeCompare(a.end_date || a.start_date);
      }
      return bTime - aTime;
    })[0];
  }, [tournamentsData]);

  const { data: leaderboardData, isLoading: leaderboardLoading } = useLeaderboard(
    recentCompleted?.id ?? "",
  );
  const { data: picksData } = useLeaguePicks(leagueId, recentCompleted?.id ?? "");

  const userPickGolferId = useMemo(() => {
    if (!currentUser || !picksData?.entries) return null;
    const pick = picksData.entries.find((p) => p.user_id === currentUser.id);
    return pick?.golfer_id ?? null;
  }, [currentUser, picksData]);

  const isLoading = tournamentsLoading || leaderboardLoading;

  if (!recentCompleted && !tournamentsLoading) {
    return null;
  }

  const allEntries = leaderboardData?.entries ?? [];
  const top5 = allEntries.slice(0, 5);
  const userPickEntry = userPickGolferId
    ? allEntries.find((e) => e.golfer_id === userPickGolferId)
    : null;
  const userPickInTop5 = userPickEntry && userPickEntry.position <= 5;

  const displayEntries = userPickInTop5 || !userPickEntry ? top5 : allEntries.slice(0, 4);
  const showUserBubble = userPickEntry && !userPickInTop5;

  const action = recentCompleted && (
    <Button asChild variant="ghost" size="sm">
      <Link href={`/${leagueId}/tournaments/${recentCompleted.id}`}>View all</Link>
    </Button>
  );

  const title = recentCompleted ? `${recentCompleted.name} Results` : "Recent Results";

  return (
    <DashboardCard
      title={title}
      icon={<TrophyIcon className="size-4" />}
      action={action}
      isLoading={isLoading}
    >
      {displayEntries.length > 0 ? (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-12">Pos</TableHead>
              <TableHead>Player</TableHead>
              <TableHead className="text-right">Score</TableHead>
              <TableHead className="text-right">Earnings</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {displayEntries.map((entry) => {
              const isUserPick = entry.golfer_id === userPickGolferId;
              return (
                <TableRow key={entry.golfer_id} className={cn(isUserPick && "bg-primary/5")}>
                  <TableCell className="font-medium">{entry.position_display}</TableCell>
                  <TableCell className="flex items-center gap-2">
                    <span className="truncate">{entry.golfer_name}</span>
                    {isUserPick && (
                      <Badge variant="outline" className="text-xs">
                        Pick
                      </Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right font-mono">
                    {formatScoreToPar(entry.score)}
                  </TableCell>
                  <TableCell className="text-right">{formatEarnings(entry.earnings)}</TableCell>
                </TableRow>
              );
            })}
            {showUserBubble && (
              <>
                <TableRow>
                  <TableCell colSpan={4} className="py-1 text-center text-xs text-muted-foreground">
                    ...
                  </TableCell>
                </TableRow>
                <TableRow className="bg-primary/5">
                  <TableCell className="font-medium">{userPickEntry.position_display}</TableCell>
                  <TableCell className="flex items-center gap-2">
                    <span className="truncate">{userPickEntry.golfer_name}</span>
                    <Badge variant="outline" className="text-xs">
                      Pick
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right font-mono">
                    {formatScoreToPar(userPickEntry.score)}
                  </TableCell>
                  <TableCell className="text-right">
                    {formatEarnings(userPickEntry.earnings)}
                  </TableCell>
                </TableRow>
              </>
            )}
          </TableBody>
        </Table>
      ) : (
        <Empty className="border-none py-4">
          <EmptyHeader>
            <EmptyTitle>Results pending</EmptyTitle>
            <EmptyDescription>
              Final standings will be posted after the tournament concludes.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      )}
    </DashboardCard>
  );
}
