"use client";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/shadcn/tabs";
import { useBreadcrumbs } from "@/components/core/breadcrumbs";
import { LeagueLeaderboard } from "@/features/leaderboards/components";
import {
  ActionLog,
  CommissionerPanel,
  LeagueMonogram,
  LeagueTournamentList,
} from "@/features/leagues/components";
import { useLeague } from "@/features/leagues/queries";
import { useEffect } from "react";

interface LeagueDetailProps {
  id: string;
}

export function LeagueDetail({ id }: LeagueDetailProps) {
  const { data, isLoading, error } = useLeague(id);
  const { setExtraCrumbs } = useBreadcrumbs();

  const leagueName = data?.league.name.trim();

  useEffect(() => {
    setExtraCrumbs(leagueName ? [{ name: leagueName }] : []);
  }, [leagueName, setExtraCrumbs]);

  useEffect(() => {
    return () => {
      setExtraCrumbs([]);
    };
  }, [setExtraCrumbs]);

  if (isLoading) {
    return (
      <div className="container mx-auto p-8">
        <p className="text-muted-foreground">Loading league...</p>
      </div>
    );
  }

  if (error || !data?.league) {
    return (
      <div className="container mx-auto p-8">
        <h1 className="mb-2 text-3xl font-bold">League Not Found</h1>
        <p className="text-muted-foreground">
          The league you&apos;re looking for doesn&apos;t exist or couldn&apos;t be loaded.
        </p>
      </div>
    );
  }

  const league = data.league;
  const isOwner = league.role === "owner";

  return (
    <div className="container mx-auto p-8">
      <div className="mb-8 flex items-center gap-4">
        <LeagueMonogram league={league} size={48} />
        <div>
          <h1 className="text-3xl font-bold">{league.name}</h1>
          <p className="text-muted-foreground">
            Season {league.season_year} &middot; {league.member_count ?? 0}{" "}
            {league.member_count === 1 ? "member" : "members"}
          </p>
        </div>
      </div>

      <Tabs defaultValue="tournaments" className="space-y-6">
        <TabsList>
          <TabsTrigger value="tournaments">Tournaments</TabsTrigger>
          <TabsTrigger value="leaderboard">Leaderboard</TabsTrigger>
          {isOwner && <TabsTrigger value="manage">Manage</TabsTrigger>}
        </TabsList>

        <TabsContent value="tournaments">
          <LeagueTournamentList leagueId={id} />
        </TabsContent>

        <TabsContent value="leaderboard">
          <LeagueLeaderboard leagueId={id} />
        </TabsContent>

        {isOwner && (
          <TabsContent value="manage" className="space-y-6">
            <CommissionerPanel league={league} />
            <ActionLog leagueId={id} />
          </TabsContent>
        )}
      </Tabs>
    </div>
  );
}
