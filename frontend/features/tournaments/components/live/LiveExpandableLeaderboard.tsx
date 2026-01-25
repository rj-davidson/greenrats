"use client";

import { GolfScorecard } from "../GolfScorecard";
import {
  formatScoreToPar,
  formatThru,
  getCurrentRoundScore,
  getRoundLabel,
} from "../leaderboard-utils";
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
import { useLeaderboard, usePrefetchLeaderboardWithHoles } from "@/features/tournaments/queries";
import type { LeaderboardEntry } from "@/features/tournaments/types";
import { cn } from "@/lib/utils";
import { ChevronDownIcon, ChevronRightIcon, UserCheck } from "lucide-react";
import { Fragment, useCallback, useState } from "react";

interface LiveExpandableLeaderboardProps {
  tournamentId: string;
  leagueId?: string;
  highlightedGolferId?: string;
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

export function LiveExpandableLeaderboard({
  tournamentId,
  leagueId,
  highlightedGolferId,
}: LiveExpandableLeaderboardProps) {
  const [expandedGolferId, setExpandedGolferId] = useState<string | null>(null);

  const { data, isLoading, error } = useLeaderboard(tournamentId, { leagueId });
  const { data: dataWithHoles } = useLeaderboard(tournamentId, { include: "holes", leagueId });
  const prefetch = usePrefetchLeaderboardWithHoles(tournamentId, leagueId);

  const toggleExpand = useCallback((golferId: string) => {
    setExpandedGolferId((prev) => (prev === golferId ? null : golferId));
  }, []);

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

  const currentRound = data.current_round;

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-12"></TableHead>
          <TableHead className="w-16">Pos</TableHead>
          <TableHead>Player</TableHead>
          <TableHead className="w-16">{getRoundLabel(currentRound)}</TableHead>
          <TableHead className="w-16">Thru</TableHead>
          <TableHead className="w-20">Total</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {data.entries.map((entry) => {
          const isExpanded = expandedGolferId === entry.golfer_id;
          const entryWithHoles = dataWithHoles?.entries.find(
            (e) => e.golfer_id === entry.golfer_id,
          );

          return (
            <Fragment key={entry.golfer_id}>
              <LiveLeaderboardRow
                entry={entry}
                isExpanded={isExpanded}
                isHighlighted={entry.golfer_id === highlightedGolferId}
                onToggle={() => toggleExpand(entry.golfer_id)}
                onHover={prefetch}
              />
              {isExpanded && entryWithHoles && (
                <TableRow className="hover:bg-transparent">
                  <TableCell colSpan={6} className="px-2 py-0">
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
                      rounds={entryWithHoles.rounds}
                      onClose={() => setExpandedGolferId(null)}
                    />
                  </TableCell>
                </TableRow>
              )}
              {isExpanded && !entryWithHoles && (
                <TableRow className="hover:bg-transparent">
                  <TableCell colSpan={6} className="p-4">
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

interface LiveLeaderboardRowProps {
  entry: LeaderboardEntry;
  isExpanded: boolean;
  isHighlighted: boolean;
  onToggle: () => void;
  onHover: () => void;
}

function LiveLeaderboardRow({
  entry,
  isExpanded,
  isHighlighted,
  onToggle,
  onHover,
}: LiveLeaderboardRowProps) {
  return (
    <TableRow
      className={cn(
        "cursor-pointer",
        isHighlighted && "bg-primary/20 hover:bg-primary/25",
      )}
      onClick={onToggle}
      onPointerEnter={onHover}
    >
      <TableCell className="text-muted-foreground">
        {isExpanded ? (
          <ChevronDownIcon className="h-4 w-4" />
        ) : (
          <ChevronRightIcon className="h-4 w-4" />
        )}
      </TableCell>
      <TableCell className={cn("font-medium", isHighlighted && "font-bold")}>
        {entry.position_display}
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
      <TableCell className="font-mono">{getCurrentRoundScore(entry)}</TableCell>
      <TableCell className="text-muted-foreground">
        {formatThru(entry.thru, entry.status)}
      </TableCell>
      <TableCell className={cn("font-mono", entry.score < 0 && "text-primary")}>
        {formatScoreToPar(entry.score)}
      </TableCell>
    </TableRow>
  );
}
