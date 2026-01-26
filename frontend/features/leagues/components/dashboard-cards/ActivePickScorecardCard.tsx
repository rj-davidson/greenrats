"use client";

import { DashboardCard } from "./DashboardCard";
import { Badge } from "@/components/shadcn/badge";
import { Button } from "@/components/shadcn/button";
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/shadcn/empty";
import { useLeagueTournaments } from "@/features/leagues/queries";
import { useLeaguePicks } from "@/features/picks/queries";
import {
  formatScoreToPar,
  getHoleScoreClass,
} from "@/features/tournaments/components/leaderboard-utils";
import { useLeaderboard } from "@/features/tournaments/queries";
import type { HoleScore, RoundScore } from "@/features/tournaments/types";
import { useCurrentUser } from "@/features/users/queries";
import { cn } from "@/lib/utils";
import { FileTextIcon } from "lucide-react";
import Link from "next/link";
import { useMemo } from "react";

interface ActivePickScorecardCardProps {
  leagueId: string;
}

function getParForHole(rounds: RoundScore[], holeNumber: number): number {
  for (const round of rounds) {
    const hole = round.holes?.find((h) => h.hole_number === holeNumber);
    if (hole) return hole.par;
  }
  return 4;
}

function getFrontNinePar(rounds: RoundScore[]): number {
  const roundWithHoles = rounds.find((r) => r.holes && r.holes.length >= 9);
  if (!roundWithHoles?.holes) return 36;
  return roundWithHoles.holes.filter((h) => h.hole_number <= 9).reduce((sum, h) => sum + h.par, 0);
}

function getBackNinePar(rounds: RoundScore[]): number {
  const roundWithHoles = rounds.find((r) => r.holes && r.holes.length === 18);
  if (!roundWithHoles?.holes) return 36;
  return roundWithHoles.holes.filter((h) => h.hole_number > 9).reduce((sum, h) => sum + h.par, 0);
}

function getTotalPar(rounds: RoundScore[]): number {
  const roundWithHoles = rounds.find((r) => r.holes && r.holes.length === 18);
  if (!roundWithHoles?.holes) return 72;
  return roundWithHoles.holes.reduce((sum, h) => sum + h.par, 0);
}

function calculateNineStrokes(holes: HoleScore[], isBack: boolean): number | null {
  const filtered = holes.filter((h) => (isBack ? h.hole_number > 9 : h.hole_number <= 9));
  const played = filtered.filter((h) => h.score !== null);
  if (played.length === 0) return null;
  return played.reduce((sum, h) => sum + (h.score ?? 0), 0);
}

export function ActivePickScorecardCard({ leagueId }: ActivePickScorecardCardProps) {
  const { data: tournamentsData, isLoading: tournamentsLoading } = useLeagueTournaments(leagueId);
  const { data: currentUser } = useCurrentUser();

  const activeTournament = useMemo(() => {
    if (!tournamentsData?.tournaments) return null;
    return tournamentsData.tournaments.find((t) => t.status === "active") ?? null;
  }, [tournamentsData]);

  const { data: picksData, isLoading: picksLoading } = useLeaguePicks(
    leagueId,
    activeTournament?.id ?? "",
  );

  const { data: leaderboardData, isLoading: leaderboardLoading } = useLeaderboard(
    activeTournament?.id ?? "",
    { include: "holes" },
  );

  const userPick = useMemo(() => {
    if (!currentUser || !picksData?.entries) return null;
    return picksData.entries.find((p) => p.user_id === currentUser.id) ?? null;
  }, [currentUser, picksData]);

  const golferData = useMemo(() => {
    if (!userPick || !leaderboardData?.entries) return null;
    return leaderboardData.entries.find((e) => e.golfer_id === userPick.golfer_id) ?? null;
  }, [userPick, leaderboardData]);

  const isLoading = tournamentsLoading || picksLoading || leaderboardLoading;

  if (!activeTournament && !tournamentsLoading) {
    return null;
  }

  if (!userPick && !isLoading) {
    return null;
  }

  const action = activeTournament && (
    <Button asChild variant="ghost" size="sm">
      <Link href={`/${leagueId}/tournaments/${activeTournament.id}`}>View</Link>
    </Button>
  );

  const frontNine = [1, 2, 3, 4, 5, 6, 7, 8, 9];
  const backNine = [10, 11, 12, 13, 14, 15, 16, 17, 18];
  const sortedRounds = golferData?.rounds
    ? [...golferData.rounds].sort((a, b) => a.round_number - b.round_number)
    : [];

  const cellClass = "px-1.5 py-0.5 text-center font-mono text-xs";
  const headerClass = cn(cellClass, "bg-muted font-semibold");

  return (
    <DashboardCard
      title="Your Pick Scorecard"
      icon={<FileTextIcon className="size-4" />}
      action={action}
      isLoading={isLoading}
    >
      {userPick && golferData ? (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-semibold">{userPick.golfer_name}</p>
              <p className="text-sm text-muted-foreground">{golferData.position_display}</p>
            </div>
            <Badge variant="outline" className="font-mono">
              {formatScoreToPar(golferData.score)}
            </Badge>
          </div>

          {sortedRounds.length > 0 && sortedRounds[0].holes ? (
            <div className="overflow-x-auto rounded border">
              <table className="w-full border-collapse text-xs">
                <thead>
                  <tr className="border-b">
                    <th className={cn(headerClass, "sticky left-0 z-10 bg-muted text-left")}>
                      Hole
                    </th>
                    {frontNine.map((h) => (
                      <th key={h} className={headerClass}>
                        {h}
                      </th>
                    ))}
                    <th className={cn(headerClass, "bg-muted/80")}>OUT</th>
                  </tr>
                  <tr className="border-b bg-muted/30">
                    <td className={cn(cellClass, "sticky left-0 z-10 bg-muted/30 text-left")}>
                      Par
                    </td>
                    {frontNine.map((h) => (
                      <td key={h} className={cellClass}>
                        {getParForHole(sortedRounds, h)}
                      </td>
                    ))}
                    <td className={cn(cellClass, "bg-muted/50 font-medium")}>
                      {getFrontNinePar(sortedRounds)}
                    </td>
                  </tr>
                </thead>
                <tbody>
                  {sortedRounds.map((round) => {
                    const getHole = (n: number) => round.holes?.find((h) => h.hole_number === n);
                    const frontStrokes = round.holes
                      ? calculateNineStrokes(round.holes, false)
                      : null;
                    return (
                      <tr key={round.round_number} className="border-b last:border-b-0">
                        <td className={cn(cellClass, "sticky left-0 z-10 bg-background text-left")}>
                          R{round.round_number}
                        </td>
                        {frontNine.map((n) => {
                          const hole = getHole(n);
                          return (
                            <td
                              key={n}
                              className={cn(
                                cellClass,
                                hole && getHoleScoreClass(hole.score, hole.par),
                              )}
                            >
                              {hole?.score ?? "-"}
                            </td>
                          );
                        })}
                        <td className={cn(cellClass, "bg-muted/30 font-medium")}>
                          {frontStrokes ?? "-"}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>

              <table className="w-full border-collapse border-t text-xs">
                <thead>
                  <tr className="border-b">
                    <th className={cn(headerClass, "sticky left-0 z-10 bg-muted text-left")}>
                      Hole
                    </th>
                    {backNine.map((h) => (
                      <th key={h} className={headerClass}>
                        {h}
                      </th>
                    ))}
                    <th className={cn(headerClass, "bg-muted/80")}>IN</th>
                    <th className={cn(headerClass, "bg-muted/80")}>TOT</th>
                  </tr>
                  <tr className="border-b bg-muted/30">
                    <td className={cn(cellClass, "sticky left-0 z-10 bg-muted/30 text-left")}>
                      Par
                    </td>
                    {backNine.map((h) => (
                      <td key={h} className={cellClass}>
                        {getParForHole(sortedRounds, h)}
                      </td>
                    ))}
                    <td className={cn(cellClass, "bg-muted/50 font-medium")}>
                      {getBackNinePar(sortedRounds)}
                    </td>
                    <td className={cn(cellClass, "bg-muted/50 font-medium")}>
                      {getTotalPar(sortedRounds)}
                    </td>
                  </tr>
                </thead>
                <tbody>
                  {sortedRounds.map((round) => {
                    const getHole = (n: number) => round.holes?.find((h) => h.hole_number === n);
                    const frontStrokes = round.holes
                      ? calculateNineStrokes(round.holes, false)
                      : null;
                    const backStrokes = round.holes
                      ? calculateNineStrokes(round.holes, true)
                      : null;
                    const totalStrokes =
                      frontStrokes !== null && backStrokes !== null
                        ? frontStrokes + backStrokes
                        : null;
                    return (
                      <tr key={round.round_number} className="border-b last:border-b-0">
                        <td className={cn(cellClass, "sticky left-0 z-10 bg-background text-left")}>
                          R{round.round_number}
                        </td>
                        {backNine.map((n) => {
                          const hole = getHole(n);
                          return (
                            <td
                              key={n}
                              className={cn(
                                cellClass,
                                hole && getHoleScoreClass(hole.score, hole.par),
                              )}
                            >
                              {hole?.score ?? "-"}
                            </td>
                          );
                        })}
                        <td className={cn(cellClass, "bg-muted/30 font-medium")}>
                          {backStrokes ?? "-"}
                        </td>
                        <td className={cn(cellClass, "bg-muted/30 font-medium")}>
                          {totalStrokes ?? "-"}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          ) : (
            <Empty className="border-none py-4">
              <EmptyHeader>
                <EmptyTitle>Waiting for round data</EmptyTitle>
                <EmptyDescription>
                  Hole-by-hole scores will appear once play begins.
                </EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}
        </div>
      ) : (
        <Empty className="border-none py-4">
          <EmptyHeader>
            <EmptyTitle>No scorecard available</EmptyTitle>
            <EmptyDescription>
              Hole-by-hole scores will appear once play begins.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      )}
    </DashboardCard>
  );
}
