"use client";

import { Button } from "@/components/shadcn/button";
import { Separator } from "@/components/shadcn/separator";
import { Skeleton } from "@/components/shadcn/skeleton";
import { LeagueCard } from "@/features/dashboard/components/LeagueCard";
import { PendingActions } from "@/features/dashboard/components/PendingActions";
import { PrimaryTournamentCard } from "@/features/dashboard/components/PrimaryTournamentCard";
import { QuickJoinInput } from "@/features/dashboard/components/QuickJoinInput";
import { CreateLeagueDialog } from "@/features/leagues/components/CreateLeagueDialog";
import { useUserLeagues } from "@/features/leagues/queries";
import { LeaderboardTable } from "@/features/tournaments/components/LeaderboardTable";
import { UserPicksByLeague } from "@/features/tournaments/components/UserPicksByLeague";
import { useCurrentTournament } from "@/features/tournaments/queries";
import { UsersIcon } from "lucide-react";
import Link from "next/link";

function TournamentSection() {
  const { tournament, isLoading } = useCurrentTournament();

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-6 w-32" />
        <Skeleton className="h-48 w-full" />
      </div>
    );
  }

  if (!tournament || tournament.status === "upcoming") {
    return null;
  }

  return (
    <>
      <UserPicksByLeague tournamentId={tournament.id} />
      <section className="space-y-3">
        <h2 className="text-lg font-semibold">Leaderboard</h2>
        <LeaderboardTable tournamentId={tournament.id} limit={6} />
      </section>
    </>
  );
}

export function DashboardView() {
  const { data: leaguesData, isLoading: leaguesLoading } = useUserLeagues();

  return (
    <div className="w-full max-w-full space-y-6 overflow-hidden p-4">
      <PrimaryTournamentCard />

      <Separator />

      <TournamentSection />

      <PendingActions />

      <section className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Your Leagues</h2>
          <div className="flex items-center gap-2">
            <QuickJoinInput />
            <CreateLeagueDialog />
          </div>
        </div>
        {leaguesLoading ? (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            <Skeleton className="h-32 w-full" />
            <Skeleton className="h-32 w-full" />
            <Skeleton className="h-32 w-full" />
          </div>
        ) : leaguesData?.leagues.length ? (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {leaguesData.leagues.map((league) => (
              <LeagueCard key={league.id} league={league} />
            ))}
          </div>
        ) : (
          <div className="rounded-lg border border-dashed p-8 text-center text-muted-foreground">
            <p className="mb-4">You haven&apos;t joined any leagues yet.</p>
            <div className="flex justify-center gap-2">
              <CreateLeagueDialog />
              <Button variant="outline" asChild>
                <Link href="/leagues/join">
                  <UsersIcon className="mr-2 size-4" />
                  Join a League
                </Link>
              </Button>
            </div>
          </div>
        )}
      </section>
    </div>
  );
}
