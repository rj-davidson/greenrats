"use client";

import { useBreadcrumbs } from "@/components/core/breadcrumbs";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/shadcn/card";
import { Skeleton } from "@/components/shadcn/skeleton";
import { LeagueLeaderboard } from "@/features/leaderboards/components";
import { useLeagueLeaderboard } from "@/features/leaderboards/queries";
import { LeagueMonogram, LeagueTournamentCard, PickWindowAlert } from "@/features/leagues/components";
import { useLeague, useLeagueTournaments } from "@/features/leagues/queries";
import type { LeagueTournament } from "@/features/leagues/types";
import { PickMaker } from "@/features/picks/components/PickMaker";
import { useLeaguePicks } from "@/features/picks/queries";
import { useTournament } from "@/features/tournaments/queries";
import { useCurrentUser } from "@/features/users/queries";
import { TrophyIcon, UsersIcon } from "lucide-react";
import { useParams } from "next/navigation";
import { useEffect, useMemo } from "react";

function formatEarnings(amount: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(amount);
}

function QuickStats({ leagueId }: { leagueId: string }) {
  const { data: currentUser } = useCurrentUser();
  const { data: leaderboardData, isLoading } = useLeagueLeaderboard(leagueId);

  if (isLoading) {
    return <Skeleton className="h-24 w-full" />;
  }

  const userEntry = leaderboardData?.entries.find((e) => e.user_id === currentUser?.id);

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium">Your Stats</CardTitle>
      </CardHeader>
      <CardContent>
        {userEntry ? (
          <div className="flex gap-6">
            <div className="flex items-center gap-2">
              <TrophyIcon className="size-4 text-muted-foreground" />
              <span className="text-2xl font-bold">{userEntry.rank}</span>
              <span className="text-sm text-muted-foreground">
                of {leaderboardData?.entries.length ?? 0}
              </span>
            </div>
            <div>
              <span className="text-2xl font-bold">{formatEarnings(userEntry.earnings)}</span>
              <span className="ml-1 text-sm text-muted-foreground">earnings</span>
            </div>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">Make your first pick to see your stats!</p>
        )}
      </CardContent>
    </Card>
  );
}

function UpcomingTournamentPick({
  leagueId,
  tournament,
}: {
  leagueId: string;
  tournament: LeagueTournament;
}) {
  const { data: tournamentData, isLoading: tournamentLoading } = useTournament(tournament.id);
  const { data: picksData } = useLeaguePicks(leagueId, tournament.id);
  const { data: currentUser } = useCurrentUser();

  const currentUserPick = useMemo(() => {
    if (!currentUser || !picksData?.entries) return undefined;
    const entry = picksData.entries.find((p) => p.user_id === currentUser.id);
    if (!entry) return undefined;
    return {
      id: entry.pick_id,
      user_id: entry.user_id,
      golfer_id: entry.golfer_id,
      golfer_name: entry.golfer_name,
      tournament_id: tournament.id,
      league_id: leagueId,
      season_year: 0,
      created_at: entry.created_at,
    };
  }, [currentUser, picksData, tournament.id, leagueId]);

  if (tournamentLoading || !tournamentData?.tournament) {
    return <Skeleton className="h-64 w-full" />;
  }

  return (
    <PickMaker leagueId={leagueId} tournament={tournamentData.tournament} currentPick={currentUserPick} />
  );
}

export default function LeagueDashboardPage() {
  const params = useParams<{ leagueId: string }>();
  const leagueId = params.leagueId;

  const { data: leagueData, isLoading: leagueLoading } = useLeague(leagueId);
  const { data: tournamentsData, isLoading: tournamentsLoading } = useLeagueTournaments(leagueId);
  const { setExtraCrumbs } = useBreadcrumbs();

  const league = leagueData?.league;

  useEffect(() => {
    if (league?.name) {
      setExtraCrumbs([{ name: league.name }]);
    }
    return () => setExtraCrumbs([]);
  }, [league?.name, setExtraCrumbs]);

  const currentTournament = useMemo(() => {
    if (!tournamentsData?.tournaments) return null;
    const active = tournamentsData.tournaments.find((t) => t.status === "active");
    if (active) return active;
    const upcoming = tournamentsData.tournaments
      .filter((t) => t.status === "upcoming")
      .sort((a, b) => new Date(a.start_date).getTime() - new Date(b.start_date).getTime());
    return upcoming[0] ?? null;
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

      <PickWindowAlert leagueId={leagueId} />

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="space-y-6 lg:col-span-2">
          {tournamentsLoading ? (
            <Skeleton className="h-64 w-full" />
          ) : currentTournament ? (
            <div className="space-y-4">
              <LeagueTournamentCard
                tournament={currentTournament}
                leagueId={leagueId}
                variant={currentTournament.status === "active" ? "live" : "next"}
              />
              {currentTournament.status === "upcoming" && (
                <UpcomingTournamentPick leagueId={leagueId} tournament={currentTournament} />
              )}
            </div>
          ) : (
            <Card>
              <CardContent className="py-12 text-center">
                <p className="text-muted-foreground">No upcoming tournaments</p>
              </CardContent>
            </Card>
          )}
        </div>

        <div className="space-y-6">
          <QuickStats leagueId={leagueId} />
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Standings</CardTitle>
            </CardHeader>
            <CardContent>
              <LeagueLeaderboard leagueId={leagueId} />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
