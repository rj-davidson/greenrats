"use client";

import { GolfScorecard } from "../GolfScorecard";
import { formatScoreToPar } from "../leaderboard-utils";
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
import { useLeaderboard, useScorecard } from "@/features/tournaments/queries";
import type { LeaderboardEntry } from "@/features/tournaments/types";
import { cn } from "@/lib/utils";
import { ChevronDownIcon, ChevronRightIcon, UserCheck } from "lucide-react";
import { Fragment, useCallback, useState } from "react";

type PlacementExpandableLeaderboardProps = {
  tournamentId: string;
  leagueId?: string;
  highlightedGolferId?: string;
};

function formatEarnings(earnings: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(earnings);
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

export function PlacementExpandableLeaderboard({
  tournamentId,
  leagueId,
  highlightedGolferId,
}: PlacementExpandableLeaderboardProps) {
  const [expandedGolferId, setExpandedGolferId] = useState<string | null>(null);

  const { data, isLoading, error } = useLeaderboard(tournamentId, { leagueId });
  const { data: scorecardData } = useScorecard(tournamentId, expandedGolferId);

  const toggleExpand = useCallback((golferId: string) => {
    setExpandedGolferId((prev) => (prev === golferId ? null : golferId));
  }, []);

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

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-6"></TableHead>
          <TableHead className="w-6">Pos</TableHead>
          <TableHead>Player</TableHead>
          <TableHead className="w-6">Total</TableHead>
          <TableHead className="w-28 text-right">Earnings</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {data.entries.map((entry) => {
          const isExpanded = expandedGolferId === entry.golfer_id;

          return (
            <Fragment key={entry.golfer_id}>
              <PlacementLeaderboardRow
                entry={entry}
                isExpanded={isExpanded}
                isHighlighted={entry.golfer_id === highlightedGolferId}
                onToggle={() => toggleExpand(entry.golfer_id)}
              />
              {isExpanded && scorecardData && (
                <TableRow className="hover:bg-transparent">
                  <TableCell colSpan={5} className="px-2 py-0">
                    {entry.picked_by && entry.picked_by.length > 0 && (
                      <div className="px-2 pt-2 text-sm text-muted-foreground">
                        <span>Picked by:</span>
                        <span className="ml-2 inline-flex flex-wrap gap-2">
                          {entry.picked_by.map((picker) => (
                            <Badge key={picker} variant="outline">
                              {picker}
                            </Badge>
                          ))}
                        </span>
                      </div>
                    )}
                    <GolfScorecard
                      rounds={scorecardData.rounds}
                      onClose={() => setExpandedGolferId(null)}
                    />
                  </TableCell>
                </TableRow>
              )}
              {isExpanded && !scorecardData && (
                <TableRow className="hover:bg-transparent">
                  <TableCell colSpan={5} className="p-4">
                    <div className="flex justify-center">
                      <Skeleton className="h-32 w-full max-w-2xl" />
                    </div>
                  </TableCell>
                </TableRow>
              )}
            </Fragment>
          );
        })}
      </TableBody>
    </Table>
  );
}

interface PlacementLeaderboardRowProps {
  entry: LeaderboardEntry;
  isExpanded: boolean;
  isHighlighted: boolean;
  onToggle: () => void;
}

function PlacementLeaderboardRow({
  entry,
  isExpanded,
  isHighlighted,
  onToggle,
}: PlacementLeaderboardRowProps) {
  const isCut = entry.status === "cut";
  const isWithdrawn = entry.status === "withdrawn";

  return (
    <TableRow
      className={cn(
        "cursor-pointer",
        (isCut || isWithdrawn) && "text-muted-foreground",
        isHighlighted && "bg-primary/20 hover:bg-primary/25",
      )}
      onClick={onToggle}
    >
      <TableCell className="text-muted-foreground">
        {isExpanded ? (
          <ChevronDownIcon className="h-4 w-4" />
        ) : (
          <ChevronRightIcon className="h-4 w-4" />
        )}
      </TableCell>
      <TableCell className={cn("font-medium", isHighlighted && "font-bold")}>
        <div className="flex items-center gap-2">
          <span>{entry.position_display}</span>
        </div>
      </TableCell>
      <TableCell>
        <div className="flex items-center gap-2">
          <span className={cn(isHighlighted && "font-bold")}>{entry.golfer_name}</span>
          {entry.picked_by && entry.picked_by.length > 0 && (
            <Badge variant="secondary" className="gap-1">
              <UserCheck className="size-3" />
              {entry.picked_by.length}
            </Badge>
          )}
        </div>
      </TableCell>
      <TableCell className={cn("font-mono", entry.score < 0 && "text-primary")}>
        {formatScoreToPar(entry.score)}
      </TableCell>
      <TableCell className="text-right font-mono">
        {entry.earnings > 0 ? formatEarnings(entry.earnings) : "-"}
      </TableCell>
    </TableRow>
  );
}
