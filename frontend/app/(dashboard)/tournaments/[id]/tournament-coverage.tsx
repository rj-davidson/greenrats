"use client";

import { Badge } from "@/components/shadcn/badge";
import { useBreadcrumbs } from "@/components/core/breadcrumbs";
import { LeaderboardTable, UserPicksByLeague } from "@/features/tournaments/components";
import { useTournament } from "@/features/tournaments/queries";
import { useEffect } from "react";

interface TournamentCoverageProps {
  id: string;
}

export function TournamentCoverage({ id }: TournamentCoverageProps) {
  const { data, isLoading, error } = useTournament(id);
  const { setExtraCrumbs } = useBreadcrumbs();

  const tournamentName = data?.tournament.name.trim();

  useEffect(() => {
    setExtraCrumbs(tournamentName ? [{ name: tournamentName }] : []);
  }, [setExtraCrumbs, tournamentName]);

  useEffect(() => {
    return () => {
      setExtraCrumbs([]);
    };
  }, [setExtraCrumbs]);

  if (isLoading) {
    return (
      <div className="container mx-auto p-8">
        <p className="text-muted-foreground">Loading tournament...</p>
      </div>
    );
  }

  if (error || !data?.tournament) {
    return (
      <div className="container mx-auto p-8">
        <h1 className="mb-2 text-3xl font-bold">Tournament Not Found</h1>
        <p className="text-muted-foreground">
          The tournament you&apos;re looking for doesn&apos;t exist or couldn&apos;t be loaded.
        </p>
      </div>
    );
  }

  const tournament = data.tournament;

  return (
    <div className="container mx-auto p-8">
      <div className="mb-8">
        <div className="mb-2 flex items-center gap-3">
          <h1 className="text-3xl font-bold">{tournament.name}</h1>
          <StatusBadge status={tournament.status} />
        </div>
        <p className="text-muted-foreground">
          {tournament.city && tournament.state
            ? `${tournament.city}, ${tournament.state}`
            : tournament.course || "Tournament coverage coming soon"}
        </p>
      </div>
      <UserPicksByLeague tournamentId={id} />
      <div>
        <h2 className="mb-4 text-xl font-semibold">Leaderboard</h2>
        <LeaderboardTable tournamentId={id} />
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  switch (status) {
    case "active":
      return <Badge>Live</Badge>;
    case "upcoming":
      return <Badge variant="outline">Upcoming</Badge>;
    case "completed":
      return <Badge variant="secondary">Final</Badge>;
    default:
      return null;
  }
}
