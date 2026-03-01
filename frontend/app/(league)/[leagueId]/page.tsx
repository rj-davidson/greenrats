"use client";

import { Card, CardContent, CardHeader } from "@/components/shadcn/card";
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
import { useMemo } from "react";

export default function LeagueDashboardPage() {
  const params = useParams<{ leagueId: string }>();
  const leagueId = params.leagueId;

  const { data: leagueData, isLoading: leagueLoading } = useLeague(leagueId);
  const { data: tournamentsData } = useLeagueTournaments(leagueId);

  const league = leagueData?.league;

  const activeTournaments = useMemo(() => {
    return tournamentsData?.tournaments.filter((t) => t.status === "active") ?? [];
  }, [tournamentsData]);

  if (leagueLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <Skeleton className="size-12 rounded-lg" />
          <div className="space-y-2">
            <Skeleton className="h-8 w-48" />
            <Skeleton className="h-4 w-32" />
          </div>
        </div>
        <div className="grid grid-cols-[repeat(auto-fit,minmax(320px,1fr))] gap-6">
          {Array.from({ length: 10 }).map((_, i) => (
            <Card key={i} className="flex flex-col gap-3 py-4">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 px-4 pb-0">
                <Skeleton className="h-4 w-24" />
              </CardHeader>
              <CardContent className="flex-1 px-4">
                <div className="space-y-2">
                  <Skeleton className="h-4 w-full" />
                  <Skeleton className="h-4 w-3/4" />
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
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

      <div className="grid grid-cols-[repeat(auto-fit,minmax(320px,1fr))] gap-6 md:grid-cols-[repeat(auto-fit,minmax(400px,1fr))]">
        <ActionCard leagueId={leagueId} />

        {activeTournaments.map((t) => (
          <ActiveTournamentCard
            key={t.id}
            leagueId={leagueId}
            tournamentId={t.id}
            tournamentName={t.name}
          />
        ))}
        {activeTournaments.map((t) => (
          <TournamentLeaderboardCard
            key={t.id}
            leagueId={leagueId}
            tournamentId={t.id}
            tournamentName={t.name}
          />
        ))}
        {activeTournaments.map((t) => (
          <ActivePickScorecardCard
            key={t.id}
            leagueId={leagueId}
            tournamentId={t.id}
            tournamentName={t.name}
          />
        ))}

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
