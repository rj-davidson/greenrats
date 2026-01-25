"use client";

import { Badge } from "@/components/shadcn/badge";
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
import { formatScoreToPar } from "../leaderboard-utils";

interface PlacementLeaderboardTableProps {
  tournamentId: string;
  limit?: number;
}

function formatEarnings(earnings: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(earnings);
}

function PlacementLeaderboardRow({ entry }: { entry: LeaderboardEntry }) {
  const isCut = entry.status === "cut";
  const isWithdrawn = entry.status === "withdrawn";
  const isTopThree = entry.position >= 1 && entry.position <= 3 && !isCut && !isWithdrawn;

  return (
    <TableRow className={cn(isCut && "text-muted-foreground")}>
      <TableCell className={cn("font-medium", isTopThree && "font-bold text-primary")}>
        <div className="flex items-center gap-2">
          <span>{entry.position_display}</span>
          {isCut && (
            <Badge variant="outline" className="text-xs">
              CUT
            </Badge>
          )}
          {isWithdrawn && (
            <Badge variant="outline" className="text-xs">
              WD
            </Badge>
          )}
        </div>
      </TableCell>
      <TableCell>
        <div className="flex items-center gap-2">
          <span className={cn(isTopThree && "font-semibold")}>{entry.golfer_name}</span>
        </div>
      </TableCell>
      <TableCell className="text-muted-foreground">{entry.country_code}</TableCell>
      <TableCell className={cn("font-mono", entry.score < 0 && "text-primary")}>
        {formatScoreToPar(entry.score)}
      </TableCell>
      <TableCell className="text-right font-mono">
        {entry.earnings > 0 ? formatEarnings(entry.earnings) : "-"}
      </TableCell>
    </TableRow>
  );
}

function PlacementLeaderboardSkeleton() {
  return (
    <div className="space-y-2">
      {Array.from({ length: 10 }).map((_, i) => (
        <Skeleton key={i} className="h-12 w-full" />
      ))}
    </div>
  );
}

export function PlacementLeaderboardTable({ tournamentId, limit }: PlacementLeaderboardTableProps) {
  const { data, isLoading, error } = useLeaderboard(tournamentId);

  if (isLoading) {
    return <PlacementLeaderboardSkeleton />;
  }

  if (error) {
    return (
      <div className="rounded-lg border border-dashed p-8 text-center">
        <p className="text-muted-foreground">Failed to load results. Please try again.</p>
      </div>
    );
  }

  if (!data || data.entries.length === 0) {
    return (
      <div className="rounded-lg border border-dashed p-8 text-center">
        <p className="text-muted-foreground">No results available yet.</p>
      </div>
    );
  }

  const entries = limit ? data.entries.slice(0, limit) : data.entries;

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-24">Pos</TableHead>
          <TableHead>Player</TableHead>
          <TableHead className="w-20">Country</TableHead>
          <TableHead className="w-20">Total</TableHead>
          <TableHead className="w-28 text-right">Earnings</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {entries.map((entry) => (
          <PlacementLeaderboardRow key={entry.golfer_id} entry={entry} />
        ))}
      </TableBody>
    </Table>
  );
}
