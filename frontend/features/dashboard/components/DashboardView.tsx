"use client";

import { LeagueCard } from "./LeagueCard";
import { PendingActions } from "./PendingActions";
import { QuickJoinInput } from "./QuickJoinInput";
import { TournamentCalendarRow } from "./TournamentCalendarRow";
import { Button } from "@/components/shadcn/button";
import { Skeleton } from "@/components/shadcn/skeleton";
import { useUserLeagues } from "@/features/leagues/queries";
import { UsersIcon } from "lucide-react";
import Link from "next/link";

export function DashboardView() {
  const { data: leaguesData, isLoading: leaguesLoading } = useUserLeagues();

  return (
    <div className="w-full max-w-full space-y-8 overflow-hidden p-4">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <QuickJoinInput />
      </div>

      <TournamentCalendarRow />

      <PendingActions />

      <section className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Your Leagues</h2>
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
    </div>
  );
}
