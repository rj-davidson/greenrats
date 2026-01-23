"use client";

import { Button } from "@/components/shadcn/button";
import { Skeleton } from "@/components/shadcn/skeleton";
import { LeagueCard, QuickJoinInput } from "@/features/dashboard/components";
import { useUserLeagues } from "@/features/leagues/queries";
import { PlusIcon, UsersIcon } from "lucide-react";
import Link from "next/link";

interface LeaguePickerContentProps {
  displayName: string;
}

export function LeaguePickerContent({ displayName }: LeaguePickerContentProps) {
  const { data: leaguesData, isLoading } = useUserLeagues();

  return (
    <div className="mx-auto max-w-4xl space-y-8">
      <div className="text-center">
        <h1 className="text-3xl font-bold">Welcome back, {displayName}!</h1>
        <p className="mt-2 text-muted-foreground">Select a league to get started</p>
      </div>

      <div className="flex flex-col items-center gap-4 sm:flex-row sm:justify-center">
        <div className="flex items-center gap-2">
          <UsersIcon className="size-5 text-muted-foreground" />
          <span className="text-sm text-muted-foreground">Join with code:</span>
          <QuickJoinInput />
        </div>
        <span className="hidden text-muted-foreground sm:inline">or</span>
        <Button asChild>
          <Link href="/create">
            <PlusIcon className="mr-2 size-4" />
            Create League
          </Link>
        </Button>
      </div>

      <div>
        <h2 className="mb-4 text-lg font-semibold">Your Leagues</h2>
        {isLoading ? (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <Skeleton className="h-32 w-full" />
            <Skeleton className="h-32 w-full" />
            <Skeleton className="h-32 w-full" />
          </div>
        ) : leaguesData?.leagues.length ? (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {leaguesData.leagues.map((league) => (
              <LeagueCard key={league.id} league={league} />
            ))}
          </div>
        ) : (
          <div className="rounded-lg border border-dashed p-12 text-center">
            <UsersIcon className="mx-auto mb-4 size-12 text-muted-foreground" />
            <h3 className="mb-2 text-lg font-medium">No leagues yet</h3>
            <p className="mb-6 text-muted-foreground">
              Create a new league or join an existing one to start competing with friends.
            </p>
            <div className="flex justify-center gap-4">
              <Button asChild>
                <Link href="/create">
                  <PlusIcon className="mr-2 size-4" />
                  Create League
                </Link>
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
