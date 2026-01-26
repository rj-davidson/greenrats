"use client";

import {
  formatScoreToPar,
  formatThru,
  getCurrentRoundScore,
  getRoundLabel,
} from "../leaderboard-utils";
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

interface LiveLeaderboardTableProps {
  tournamentId: string;
  limit?: number;
}

function LiveLeaderboardRow({
  entry,
  tournamentRound,
}: {
  entry: LeaderboardEntry;
  tournamentRound: number;
}) {
  const isTopThree = entry.position >= 1 && entry.position <= 3;
  const playerBehind = entry.current_round < tournamentRound;

  return (
    <TableRow>
      <TableCell className={cn("font-medium", isTopThree && "font-bold text-primary")}>
        {entry.position_display}
      </TableCell>
      <TableCell>
        <div className="flex items-center gap-2">
          <span className={cn(isTopThree && "font-semibold")}>{entry.golfer_name}</span>
        </div>
      </TableCell>
      <TableCell className="text-muted-foreground">{entry.country_code}</TableCell>
      <TableCell className="font-mono">
        {playerBehind ? "-" : getCurrentRoundScore(entry)}
      </TableCell>
      <TableCell className="text-muted-foreground">
        {playerBehind ? "-" : formatThru(entry.thru, entry.status)}
      </TableCell>
      <TableCell className={cn("font-mono", entry.score < 0 && "text-primary")}>
        {formatScoreToPar(entry.score)}
      </TableCell>
    </TableRow>
  );
}

function LiveLeaderboardSkeleton() {
  return (
    <div className="space-y-2">
      {Array.from({ length: 10 }).map((_, i) => (
        <Skeleton key={i} className="h-12 w-full" />
      ))}
    </div>
  );
}

export function LiveLeaderboardTable({ tournamentId, limit }: LiveLeaderboardTableProps) {
  const { data, isLoading, error } = useLeaderboard(tournamentId);

  if (isLoading) {
    return <LiveLeaderboardSkeleton />;
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

  const entries = limit ? data.entries.slice(0, limit) : data.entries;
  const currentRound = data.current_round;

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-16">Pos</TableHead>
          <TableHead>Player</TableHead>
          <TableHead className="w-20">Country</TableHead>
          <TableHead className="w-16">{getRoundLabel(currentRound)}</TableHead>
          <TableHead className="w-16">Thru</TableHead>
          <TableHead className="w-20">Total</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {entries.map((entry) => (
          <LiveLeaderboardRow key={entry.golfer_id} entry={entry} tournamentRound={currentRound} />
        ))}
      </TableBody>
    </Table>
  );
}
