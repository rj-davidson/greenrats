"use client";

import { useBreadcrumbs } from "@/components/core/breadcrumbs";
import { Skeleton } from "@/components/shadcn/skeleton";
import {
  ActionCard,
  ActivePickScorecardCard,
  ActiveTournamentCard,
  AuditCard,
  LeagueMonogram,
  PickHistoryCard,
  RecentTournamentResultsCard,
  SeasonProgressCard,
  StandingsCard,
  TournamentLeaderboardCard,
  YourStatsCard,
} from "@/features/leagues/components";
import { useLeague, useLeagueTournaments } from "@/features/leagues/queries";
import { UsersIcon } from "lucide-react";
import { useParams } from "next/navigation";
import { useEffect, useMemo } from "react";

export default function LeagueDashboardPage() {
  const params = useParams<{ leagueId: string }>();
  const leagueId = params.leagueId;

  const { data: leagueData, isLoading: leagueLoading } = useLeague(leagueId);
  const { data: tournamentsData } = useLeagueTournaments(leagueId);
  const { setExtraCrumbs } = useBreadcrumbs();

  const league = leagueData?.league;

  useEffect(() => {
    if (league?.name) {
      setExtraCrumbs([{ name: league.name }]);
    }
    return () => setExtraCrumbs([]);
  }, [league?.name, setExtraCrumbs]);

  const hasActiveTournament = useMemo(() => {
    return tournamentsData?.tournaments.some((t) => t.status === "active") ?? false;
  }, [tournamentsData]);

  if (leagueLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (!league) {
    return (
      <div className="text-center">
        <h1 className="mb-2 text-2xl font-bold">League Not Found</h1>
        <p className="text-muted-foreground">The league you are looking for does not exist.</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <LeagueMonogram league={league} size={48} />
        <div>
          <h1 className="text-3xl font-bold">{league.name}</h1>
          <div className="flex items-center gap-3 text-muted-foreground">
            <span>Season {league.season_year}</span>
            <span className="flex items-center gap-1">
              <UsersIcon className="size-4" />
              {league.member_count ?? 0} {(league.member_count ?? 0) === 1 ? "member" : "members"}
            </span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-[repeat(auto-fit,minmax(320px,1fr))] gap-6">
        <ActionCard leagueId={leagueId} />

        {hasActiveTournament && (
          <>
            <ActiveTournamentCard leagueId={leagueId} />
            <TournamentLeaderboardCard leagueId={leagueId} />
            <ActivePickScorecardCard leagueId={leagueId} />
          </>
        )}

        <RecentTournamentResultsCard leagueId={leagueId} />
        <YourStatsCard leagueId={leagueId} />
        <StandingsCard leagueId={leagueId} />
        <SeasonProgressCard leagueId={leagueId} />
        <PickHistoryCard leagueId={leagueId} />
        <AuditCard leagueId={leagueId} />
      </div>
    </div>
  );
}
