"use client";

import { DashboardCard } from "./DashboardCard";
import { Badge } from "@/components/shadcn/badge";
import { Button } from "@/components/shadcn/button";
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
import {
  formatScoreToPar,
  formatThru,
  getCurrentRoundScore,
  getRoundLabel,
} from "@/features/tournaments/components/leaderboard-utils";
import { PositionChangeIndicator } from "@/features/tournaments/components/live/PositionChangeIndicator";
import { useLeaderboard } from "@/features/tournaments/queries";
import { useCurrentUser } from "@/features/users/queries";
import { cn } from "@/lib/utils";
import { ListOrderedIcon } from "lucide-react";
import Link from "next/link";
import { useMemo } from "react";

interface TournamentLeaderboardCardProps {
  leagueId: string;
}

export function TournamentLeaderboardCard({ leagueId }: TournamentLeaderboardCardProps) {
  const { data: tournamentsData, isLoading: tournamentsLoading } = useLeagueTournaments(leagueId);
  const { data: currentUser } = useCurrentUser();

  const activeTournament = useMemo(() => {
    if (!tournamentsData?.tournaments) return null;
    return tournamentsData.tournaments.find((t) => t.status === "active") ?? null;
  }, [tournamentsData]);

  const { data: leaderboardData, isLoading: leaderboardLoading } = useLeaderboard(
    activeTournament?.id ?? "",
  );

  const { data: picksData } = useLeaguePicks(leagueId, activeTournament?.id ?? "");

  const userPickGolferId = useMemo(() => {
    if (!currentUser || !picksData?.entries) return null;
    const pick = picksData.entries.find((p) => p.user_id === currentUser.id);
    return pick?.golfer_id ?? null;
  }, [currentUser, picksData]);

  const isLoading = tournamentsLoading || leaderboardLoading;

  if (!activeTournament && !tournamentsLoading) {
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

  const action = activeTournament && (
    <Button asChild variant="ghost" size="sm">
      <Link href={`/${leagueId}/tournaments/${activeTournament.id}`}>View all</Link>
    </Button>
  );

  const title = activeTournament ? activeTournament.name : "Leaderboard";
  const currentRound = leaderboardData?.current_round ?? 1;
  const showPositionChange = currentRound >= 2;

  return (
    <DashboardCard
      title={title}
      icon={<ListOrderedIcon className="size-4" />}
      action={action}
      isLoading={isLoading}
    >
      {displayEntries.length > 0 ? (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-12">Pos</TableHead>
              {showPositionChange && <TableHead className="w-10">+/-</TableHead>}
              <TableHead>Player</TableHead>
              <TableHead className="text-right">{getRoundLabel(currentRound)}</TableHead>
              <TableHead className="text-right">Thru</TableHead>
              <TableHead className="text-right">Total</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {displayEntries.map((entry) => {
              const isUserPick = entry.golfer_id === userPickGolferId;
              const playerBehind = entry.current_round < currentRound;
              return (
                <TableRow key={entry.golfer_id} className={cn(isUserPick && "bg-primary/5")}>
                  <TableCell className="font-medium">{entry.position_display}</TableCell>
                  {showPositionChange && (
                    <TableCell>
                      {playerBehind ? (
                        <span className="text-muted-foreground">-</span>
                      ) : (
                        <PositionChangeIndicator change={entry.position_change} />
                      )}
                    </TableCell>
                  )}
                  <TableCell className="flex items-center gap-2">
                    <span className="truncate">{entry.golfer_name}</span>
                    {isUserPick && (
                      <Badge variant="outline" className="text-xs">
                        Pick
                      </Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right font-mono">
                    {playerBehind ? "-" : getCurrentRoundScore(entry)}
                  </TableCell>
                  <TableCell className="text-right text-muted-foreground">
                    {playerBehind ? "-" : formatThru(entry.thru, entry.status)}
                  </TableCell>
                  <TableCell className="text-right font-mono">
                    {formatScoreToPar(entry.score)}
                  </TableCell>
                </TableRow>
              );
            })}
            {showUserBubble && (
              <>
                <TableRow>
                  <TableCell
                    colSpan={showPositionChange ? 6 : 5}
                    className="py-1 text-center text-xs text-muted-foreground"
                  >
                    ...
                  </TableCell>
                </TableRow>
                {(() => {
                  const userPlayerBehind = userPickEntry.current_round < currentRound;
                  return (
                    <TableRow className="bg-primary/5">
                      <TableCell className="font-medium">{userPickEntry.position_display}</TableCell>
                      {showPositionChange && (
                        <TableCell>
                          {userPlayerBehind ? (
                            <span className="text-muted-foreground">-</span>
                          ) : (
                            <PositionChangeIndicator change={userPickEntry.position_change} />
                          )}
                        </TableCell>
                      )}
                      <TableCell className="flex items-center gap-2">
                        <span className="truncate">{userPickEntry.golfer_name}</span>
                        <Badge variant="outline" className="text-xs">
                          Pick
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right font-mono">
                        {userPlayerBehind ? "-" : getCurrentRoundScore(userPickEntry)}
                      </TableCell>
                      <TableCell className="text-right text-muted-foreground">
                        {userPlayerBehind ? "-" : formatThru(userPickEntry.thru, userPickEntry.status)}
                      </TableCell>
                      <TableCell className="text-right font-mono">
                        {formatScoreToPar(userPickEntry.score)}
                      </TableCell>
                    </TableRow>
                  );
                })()}
              </>
            )}
          </TableBody>
        </Table>
      ) : (
        <p className="text-sm text-muted-foreground">Leaderboard data not available yet.</p>
      )}
    </DashboardCard>
  );
}
