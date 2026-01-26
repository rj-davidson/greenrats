"use client";

import { Button } from "@/components/shadcn/button";
import { Skeleton } from "@/components/shadcn/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/shadcn/table";
import { useLeagueStandings } from "@/features/leaderboards/queries";
import type { StandingsEntry } from "@/features/leaderboards/types";
import { usePrefetchUserPublicPicks, useUserPublicPicks } from "@/features/picks/queries";
import { useCurrentUser } from "@/features/users/queries";
import { cn } from "@/lib/utils";
import { ChevronDownIcon, ChevronRightIcon, ChevronUpIcon } from "lucide-react";
import { Fragment, useCallback, useState } from "react";

interface ExpandableLeagueStandingsProps {
  leagueId: string;
}

function formatEarnings(amount: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(amount);
}

function StandingsSkeleton() {
  return (
    <div className="space-y-2">
      {Array.from({ length: 8 }).map((_, i) => (
        <Skeleton key={i} className="h-12 w-full" />
      ))}
    </div>
  );
}

interface ExpandedPickHistoryProps {
  leagueId: string;
  userId: string;
  onClose: () => void;
}

function ExpandedPickHistory({ leagueId, userId, onClose }: ExpandedPickHistoryProps) {
  const { data, isLoading } = useUserPublicPicks(leagueId, userId);

  const cellClass = "px-2 py-1 text-center font-mono text-xs";
  const headerCellClass = cn(cellClass, "bg-muted font-semibold");

  if (isLoading) {
    return (
      <div className="py-2">
        <div className="space-y-1">
          <Skeleton className="h-6 w-full" />
          <Skeleton className="h-6 w-full" />
          <Skeleton className="h-6 w-full" />
        </div>
      </div>
    );
  }

  if (!data?.picks.length) {
    return (
      <div className="py-2">
        <div className="rounded border px-4 py-6 text-center text-sm text-muted-foreground">
          No picks yet this season.
        </div>
        <div className="mt-2 flex justify-center">
          <Button variant="ghost" size="sm" onClick={onClose} className="w-full gap-1">
            <ChevronUpIcon className="h-4 w-4" />
          </Button>
        </div>
      </div>
    );
  }

  const totalEarnings = data.picks.reduce((sum, pick) => sum + pick.earnings, 0);

  return (
    <div className="py-2">
      <div className="overflow-x-auto rounded border">
        <table className="w-full border-collapse text-sm">
          <thead>
            <tr className="border-b">
              <th className={cn(headerCellClass, "text-left")}>Tournament</th>
              <th className={cn(headerCellClass, "text-left")}>Pick</th>
              <th className={headerCellClass}>Pos</th>
              <th className={cn(headerCellClass, "text-right")}>Earnings</th>
            </tr>
          </thead>
          <tbody>
            {data.picks.map((pick) => (
              <tr key={pick.tournament_id} className="border-b last:border-b-0">
                <td className={cn(cellClass, "text-left")}>{pick.tournament_name}</td>
                <td className={cn(cellClass, "text-left")}>{pick.golfer_name}</td>
                <td className={cellClass}>{pick.position_display || "-"}</td>
                <td className={cn(cellClass, "text-right")}>{formatEarnings(pick.earnings)}</td>
              </tr>
            ))}
            <tr className="border-t bg-muted/30">
              <td className={cn(cellClass, "text-left font-medium")} colSpan={3}>
                Total
              </td>
              <td className={cn(cellClass, "text-right font-medium")}>
                {formatEarnings(totalEarnings)}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div className="mt-2 flex justify-center">
        <Button variant="ghost" size="sm" onClick={onClose} className="w-full gap-1">
          <ChevronUpIcon className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}

interface StandingsRowProps {
  entry: StandingsEntry;
  isExpanded: boolean;
  isCurrentUser: boolean;
  showCurrentPick: boolean;
  onToggle: () => void;
  onHover: () => void;
}

function StandingsRow({
  entry,
  isExpanded,
  isCurrentUser,
  showCurrentPick,
  onToggle,
  onHover,
}: StandingsRowProps) {
  return (
    <TableRow
      className={cn("cursor-pointer", isCurrentUser && "bg-primary/20 hover:bg-primary/25")}
      onClick={onToggle}
      onPointerEnter={onHover}
    >
      <TableCell className="w-8 text-muted-foreground">
        {isExpanded ? (
          <ChevronDownIcon className="h-4 w-4" />
        ) : (
          <ChevronRightIcon className="h-4 w-4" />
        )}
      </TableCell>
      <TableCell className={cn("w-16", isCurrentUser && "font-bold")}>
        {entry.rank_display}
      </TableCell>
      <TableCell>
        <span className={cn(isCurrentUser && "font-bold")}>{entry.user_display_name}</span>
        {isCurrentUser && <span className="ml-2 text-sm text-muted-foreground">(you)</span>}
      </TableCell>
      {showCurrentPick && <TableCell>{entry.current_pick?.golfer_name ?? "--"}</TableCell>}
      <TableCell className="text-right">{formatEarnings(entry.total_earnings)}</TableCell>
    </TableRow>
  );
}

export function ExpandableLeagueStandings({ leagueId }: ExpandableLeagueStandingsProps) {
  const [expandedUserId, setExpandedUserId] = useState<string | null>(null);

  const { data, isLoading, error } = useLeagueStandings(leagueId);
  const { data: currentUser } = useCurrentUser();
  const prefetch = usePrefetchUserPublicPicks(leagueId);

  const toggleExpand = useCallback((userId: string) => {
    setExpandedUserId((prev) => (prev === userId ? null : userId));
  }, []);

  const handleHover = useCallback(
    (userId: string) => {
      prefetch(userId);
    },
    [prefetch],
  );

  if (isLoading) {
    return <StandingsSkeleton />;
  }

  if (error) {
    return <div className="text-destructive">Failed to load standings</div>;
  }

  if (!data?.entries.length) {
    return (
      <div className="py-8 text-center text-muted-foreground">
        No picks have been made yet. The standings will appear once members start making picks.
      </div>
    );
  }

  const showCurrentPick = data.active_tournament?.is_pick_window_closed ?? false;

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-8"></TableHead>
          <TableHead className="w-16">Rank</TableHead>
          <TableHead>Player</TableHead>
          {showCurrentPick && (
            <TableHead>
              Current Pick
              <span className="ml-2 text-xs font-normal text-muted-foreground">
                ({data.active_tournament?.name})
              </span>
            </TableHead>
          )}
          <TableHead className="text-right">Earnings</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {data.entries.map((entry) => {
          const isExpanded = expandedUserId === entry.user_id;
          const isCurrentUser = currentUser?.id === entry.user_id;

          return (
            <Fragment key={entry.user_id}>
              <StandingsRow
                entry={entry}
                isExpanded={isExpanded}
                isCurrentUser={isCurrentUser}
                showCurrentPick={showCurrentPick}
                onToggle={() => toggleExpand(entry.user_id)}
                onHover={() => handleHover(entry.user_id)}
              />
              {isExpanded && (
                <TableRow className="hover:bg-transparent">
                  <TableCell colSpan={showCurrentPick ? 5 : 4} className="px-2 py-0">
                    <ExpandedPickHistory
                      leagueId={leagueId}
                      userId={entry.user_id}
                      onClose={() => setExpandedUserId(null)}
                    />
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
