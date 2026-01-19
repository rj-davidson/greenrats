"use client";

import { LeagueCard } from "./LeagueCard";
import { PendingActions } from "./PendingActions";
import { UpcomingTournaments } from "./UpcomingTournamentAlert";
import { Button } from "@/components/shadcn/button";
import { Skeleton } from "@/components/shadcn/skeleton";
import { useUserLeagues } from "@/features/leagues/queries";
import { PlusIcon, UsersIcon } from "lucide-react";
import Link from "next/link";

export function DashboardView() {
  const { data: leaguesData, isLoading: leaguesLoading } = useUserLeagues();

  return (
    <div className="container mx-auto space-y-8 p-8">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <div className="flex gap-2">
          <Button asChild variant="outline">
            <Link href="/leagues/join">
              <UsersIcon className="mr-2 size-4" />
              Join League
            </Link>
          </Button>
        </div>
      </div>

      <PendingActions />

      <section className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Your Leagues</h2>
        </div>
        {leaguesLoading ? (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-20 w-full" />
          </div>
        ) : leaguesData?.leagues.length ? (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {leaguesData.leagues.map((league) => (
              <LeagueCard key={league.id} league={league} />
            ))}
          </div>
        ) : (
          <div className="text-muted-foreground rounded-lg border border-dashed p-8 text-center">
            <p className="mb-4">You haven&apos;t joined any leagues yet.</p>
            <Button asChild>
              <Link href="/leagues/join">
                <UsersIcon className="mr-2 size-4" />
                Join a League
              </Link>
            </Button>
          </div>
        )}
      </section>

      <UpcomingTournaments />
    </div>
  );
}
