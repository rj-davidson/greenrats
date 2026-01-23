"use client";

import { Input } from "@/components/shadcn/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/shadcn/table";
import type { LeagueTournament } from "@/features/leagues/types";
import { cn } from "@/lib/utils";
import { CheckCircle2Icon, SearchIcon } from "lucide-react";
import { useRouter } from "next/navigation";
import { useMemo, useState } from "react";

interface TournamentDataTableProps {
  tournaments: LeagueTournament[];
  leagueId: string;
}

function formatDateRange(startDate: string, endDate: string): string {
  const start = new Date(startDate);
  const end = new Date(endDate);
  const options: Intl.DateTimeFormatOptions = { month: "short", day: "numeric" };

  if (start.getMonth() === end.getMonth()) {
    return `${start.toLocaleDateString("en-US", options)} - ${end.getDate()}`;
  }
  return `${start.toLocaleDateString("en-US", options)} - ${end.toLocaleDateString("en-US", options)}`;
}

function formatEarnings(earnings: number): string {
  if (earnings >= 1_000_000) {
    return `$${(earnings / 1_000_000).toFixed(2)}M`;
  }
  if (earnings >= 1_000) {
    return `$${(earnings / 1_000).toFixed(0)}K`;
  }
  return `$${earnings.toLocaleString()}`;
}

function formatPickWindowDateTime(isoString: string | undefined): string {
  if (!isoString) return "--";
  const date = new Date(isoString);
  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

export function TournamentDataTable({ tournaments, leagueId }: TournamentDataTableProps) {
  const router = useRouter();
  const [search, setSearch] = useState("");

  const sortedTournaments = useMemo(() => {
    return [...tournaments].sort(
      (a, b) => new Date(a.start_date).getTime() - new Date(b.start_date).getTime(),
    );
  }, [tournaments]);

  const highlightedId = useMemo(() => {
    const active = sortedTournaments.find((t) => t.status === "active");
    if (active) return active.id;
    const firstUpcoming = sortedTournaments.find((t) => t.status === "upcoming");
    return firstUpcoming?.id ?? null;
  }, [sortedTournaments]);

  const filteredTournaments = useMemo(() => {
    if (!search.trim()) return sortedTournaments;
    const searchLower = search.toLowerCase();
    return sortedTournaments.filter((t) => t.name.toLowerCase().includes(searchLower));
  }, [sortedTournaments, search]);

  return (
    <div className="space-y-4">
      <div className="relative max-w-sm">
        <SearchIcon className="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search tournaments..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="pl-9"
        />
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Tournament</TableHead>
              <TableHead className="hidden sm:table-cell">Dates</TableHead>
              <TableHead className="hidden lg:table-cell">Picks Open</TableHead>
              <TableHead className="hidden lg:table-cell">Picks Close</TableHead>
              <TableHead>Your Pick</TableHead>
              <TableHead className="text-right">Earnings</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredTournaments.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="py-8 text-center text-muted-foreground">
                  No tournaments found
                </TableCell>
              </TableRow>
            ) : (
              filteredTournaments.map((tournament) => {
                const isHighlighted = tournament.id === highlightedId;
                const isCompleted = tournament.status === "completed";
                const hasEarnings =
                  isCompleted &&
                  tournament.golfer_earnings !== undefined &&
                  tournament.golfer_earnings > 0;

                return (
                  <TableRow
                    key={tournament.id}
                    className={cn(
                      "cursor-pointer",
                      isHighlighted && "bg-primary/20 font-medium",
                      isCompleted && "text-muted-foreground italic",
                    )}
                    onClick={() => router.push(`/${leagueId}/tournaments/${tournament.id}`)}
                  >
                    <TableCell>
                      <div className={cn("font-medium", isCompleted && "font-normal")}>
                        {tournament.name}
                      </div>
                      <div className="text-xs text-muted-foreground sm:hidden">
                        {formatDateRange(tournament.start_date, tournament.end_date)}
                      </div>
                    </TableCell>
                    <TableCell className="hidden sm:table-cell">
                      {formatDateRange(tournament.start_date, tournament.end_date)}
                    </TableCell>
                    <TableCell className="hidden lg:table-cell">
                      {formatPickWindowDateTime(tournament.pick_window_opens_at)}
                    </TableCell>
                    <TableCell className="hidden lg:table-cell">
                      {formatPickWindowDateTime(tournament.pick_window_closes_at)}
                    </TableCell>
                    <TableCell>
                      {tournament.has_user_pick ? (
                        <div
                          className={cn(
                            "flex items-center gap-1.5",
                            !isCompleted && "text-green-600",
                          )}
                        >
                          <CheckCircle2Icon
                            className={cn("size-4 shrink-0", !isCompleted && "text-green-600")}
                          />
                          <span className="truncate">{tournament.golfer_name}</span>
                        </div>
                      ) : (
                        <span>--</span>
                      )}
                    </TableCell>
                    <TableCell className="text-right">
                      {hasEarnings ? (
                        <span>{formatEarnings(tournament.golfer_earnings!)}</span>
                      ) : (
                        <span>--</span>
                      )}
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
