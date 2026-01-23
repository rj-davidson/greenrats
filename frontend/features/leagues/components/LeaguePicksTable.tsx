"use client";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/shadcn/table";
import type { LeaguePickEntry } from "@/features/picks/types";
import { CheckCircle2Icon } from "lucide-react";

interface LeaguePicksTableProps {
  picks: LeaguePickEntry[];
  tournamentStatus: string;
}

function formatEarnings(amount?: number): string {
  if (!amount) return "-";
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(amount);
}

export function LeaguePicksTable({ picks, tournamentStatus }: LeaguePicksTableProps) {
  const showGolferDetails = tournamentStatus !== "upcoming";

  const sortedPicks = [...picks].sort((a, b) => {
    if (showGolferDetails) {
      const posA = a.leaderboard?.position ?? Infinity;
      const posB = b.leaderboard?.position ?? Infinity;
      return posA - posB;
    }
    return a.user_display_name.localeCompare(b.user_display_name);
  });

  if (picks.length === 0) {
    return (
      <div className="py-8 text-center text-muted-foreground">
        No picks have been made for this tournament yet.
      </div>
    );
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Player</TableHead>
          {showGolferDetails ? (
            <>
              <TableHead>Golfer</TableHead>
              <TableHead className="text-right">Position</TableHead>
              <TableHead className="text-right">Earnings</TableHead>
            </>
          ) : (
            <TableHead>Status</TableHead>
          )}
        </TableRow>
      </TableHeader>
      <TableBody>
        {sortedPicks.map((pick) => (
          <TableRow key={pick.pick_id}>
            <TableCell className="font-medium">{pick.user_display_name || "Unknown"}</TableCell>
            {showGolferDetails ? (
              <>
                <TableCell>{pick.golfer_name}</TableCell>
                <TableCell className="text-right">
                  {pick.leaderboard?.position_display ?? "-"}
                </TableCell>
                <TableCell className="text-right">
                  {formatEarnings(pick.leaderboard?.earnings)}
                </TableCell>
              </>
            ) : (
              <TableCell>
                <div className="flex items-center gap-1.5 text-primary">
                  <CheckCircle2Icon className="size-4" />
                  Pick made
                </div>
              </TableCell>
            )}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
