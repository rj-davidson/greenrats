"use client";

import { DashboardCard } from "./DashboardCard";
import { Badge } from "@/components/shadcn/badge";
import { Button } from "@/components/shadcn/button";
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/shadcn/empty";
import { useLeagueTournaments } from "@/features/leagues/queries";
import { useLeaguePicks } from "@/features/picks/queries";
import { formatScoreToPar } from "@/features/tournaments/components/leaderboard-utils";
import { GolfScorecard } from "@/features/tournaments/components/GolfScorecard";
import { useLeaderboard, useScorecard } from "@/features/tournaments/queries";
import { useCurrentUser } from "@/features/users/queries";
import { FileTextIcon } from "lucide-react";
import Link from "next/link";
import { useMemo } from "react";

interface ActivePickScorecardCardProps {
  leagueId: string;
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

  const userPick = useMemo(() => {
    if (!currentUser || !picksData?.entries) return null;
    return picksData.entries.find((p) => p.user_id === currentUser.id) ?? null;
  }, [currentUser, picksData]);

  const { data: leaderboardData, isLoading: leaderboardLoading } = useLeaderboard(
    activeTournament?.id ?? "",
  );

  const { data: scorecardData, isLoading: scorecardLoading } = useScorecard(
    activeTournament?.id ?? "",
    userPick?.golfer_id ?? null,
  );

  const golferData = useMemo(() => {
    if (!userPick || !leaderboardData?.entries) return null;
    return leaderboardData.entries.find((e) => e.golfer_id === userPick.golfer_id) ?? null;
  }, [userPick, leaderboardData]);

  const hasRoundData = useMemo(() => {
    if (!scorecardData?.rounds) return false;
    return scorecardData.rounds.some((r) => r.holes && r.holes.length > 0);
  }, [scorecardData]);

  const isLoading = tournamentsLoading || picksLoading || leaderboardLoading || scorecardLoading;

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

          {hasRoundData && scorecardData ? (
            <GolfScorecard rounds={scorecardData.rounds} />
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
            <EmptyDescription>Hole-by-hole scores will appear once play begins.</EmptyDescription>
          </EmptyHeader>
        </Empty>
      )}
    </DashboardCard>
  );
}
