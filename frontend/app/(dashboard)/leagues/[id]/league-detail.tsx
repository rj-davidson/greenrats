"use client";

import { LeagueMonogram } from "@/features/leagues/components";
import { useLeague } from "@/features/leagues/queries";

interface LeagueDetailProps {
  id: string;
}

export function LeagueDetail({ id }: LeagueDetailProps) {
  const { data, isLoading, error } = useLeague(id);

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

  return (
    <div className="container mx-auto p-8">
      <div className="mb-8 flex items-center gap-4">
        <LeagueMonogram league={league} size={48} />
        <div>
          <h1 className="text-3xl font-bold">{league.name}</h1>
          <p className="text-muted-foreground">Season {league.season_year}</p>
        </div>
      </div>
      <div className="rounded-lg border border-dashed p-12 text-center">
        <p className="text-muted-foreground">League details coming soon</p>
      </div>
    </div>
  );
}
