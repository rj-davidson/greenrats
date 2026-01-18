"use client";

import { Skeleton } from "@/components/shadcn/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/shadcn/table";
import { useLeaderboard } from "@/features/tournaments/queries";
import type { LeaderboardEntry } from "@/features/tournaments/types";
import { cn } from "@/lib/utils";

interface LeaderboardTableProps {
  tournamentId: string;
}

function formatScore(score: number): string {
  if (score === 0) return "E";
  return score > 0 ? `+${score}` : `${score}`;
}

function formatThru(thru: number, status: string): string {
  if (status === "finished") return "F";
  if (thru === 0) return "-";
  if (thru === 18) return "F";
  return `${thru}`;
}

function LeaderboardRow({ entry }: { entry: LeaderboardEntry }) {
  const isTopThree = entry.position >= 1 && entry.position <= 3 && !entry.cut;
  const isCut = entry.cut;

  return (
    <TableRow className={cn(isCut && "text-muted-foreground")}>
      <TableCell className={cn("font-medium", isTopThree && "text-primary font-bold")}>
        {entry.position_display}
      </TableCell>
      <TableCell>
        <div className="flex items-center gap-2">
          <span className={cn(isTopThree && "font-semibold")}>{entry.golfer_name}</span>
        </div>
      </TableCell>
      <TableCell className="text-muted-foreground">{entry.country_code}</TableCell>
      <TableCell className={cn("font-mono", entry.score < 0 && "text-green-600 dark:text-green-400")}>
        {formatScore(entry.score)}
      </TableCell>
      <TableCell className="text-muted-foreground">{formatThru(entry.thru, entry.status)}</TableCell>
    </TableRow>
  );
}

function LeaderboardSkeleton() {
  return (
    <div className="space-y-2">
      {Array.from({ length: 10 }).map((_, i) => (
        <Skeleton key={i} className="h-12 w-full" />
      ))}
    </div>
  );
}

export function LeaderboardTable({ tournamentId }: LeaderboardTableProps) {
  const { data, isLoading, error } = useLeaderboard(tournamentId);

  if (isLoading) {
    return <LeaderboardSkeleton />;
  }

  if (error) {
    return (
      <div className="rounded-lg border border-dashed p-8 text-center">
        <p className="text-muted-foreground">Failed to load leaderboard. Please try again.</p>
      </div>
    );
  }

  if (!data || data.entries.length === 0) {
    return (
      <div className="rounded-lg border border-dashed p-8 text-center">
        <p className="text-muted-foreground">No leaderboard data available yet.</p>
      </div>
    );
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-16">Pos</TableHead>
          <TableHead>Player</TableHead>
          <TableHead className="w-20">Country</TableHead>
          <TableHead className="w-20">Score</TableHead>
          <TableHead className="w-16">Thru</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {data.entries.map((entry) => (
          <LeaderboardRow key={entry.golfer_id} entry={entry} />
        ))}
      </TableBody>
    </Table>
  );
}
