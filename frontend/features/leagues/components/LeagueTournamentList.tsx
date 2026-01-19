"use client";

import { useLeagueTournaments } from "@/features/leagues/queries";
import { LeagueTournamentCard } from "@/features/leagues/components/LeagueTournamentCard";
import { Skeleton } from "@/components/shadcn/skeleton";

interface LeagueTournamentListProps {
  leagueId: string;
}

export function LeagueTournamentList({ leagueId }: LeagueTournamentListProps) {
  const { data, isLoading, error } = useLeagueTournaments(leagueId);

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-32 w-full" />
      </div>
    );
  }

  if (error) {
    return <div className="text-destructive">Failed to load tournaments</div>;
  }

  if (!data || data.tournaments.length === 0) {
    return <div className="py-8 text-center text-muted-foreground">No tournaments found</div>;
  }

  const activeTournaments = data.tournaments.filter((t) => t.status === "in_progress");
  const upcomingTournaments = data.tournaments.filter((t) => t.status === "upcoming");
  const completedTournaments = data.tournaments.filter((t) => t.status === "completed");

  return (
    <div className="space-y-6">
      {activeTournaments.length > 0 && (
        <section>
          <h3 className="mb-3 text-lg font-semibold">Live</h3>
          <div className="space-y-3">
            {activeTournaments.map((tournament) => (
              <LeagueTournamentCard
                key={tournament.id}
                tournament={tournament}
                leagueId={leagueId}
              />
            ))}
          </div>
        </section>
      )}

      {upcomingTournaments.length > 0 && (
        <section>
          <h3 className="mb-3 text-lg font-semibold">Upcoming</h3>
          <div className="space-y-3">
            {upcomingTournaments.map((tournament) => (
              <LeagueTournamentCard
                key={tournament.id}
                tournament={tournament}
                leagueId={leagueId}
              />
            ))}
          </div>
        </section>
      )}

      {completedTournaments.length > 0 && (
        <section>
          <h3 className="mb-3 text-lg font-semibold">Completed</h3>
          <div className="space-y-3">
            {completedTournaments.map((tournament) => (
              <LeagueTournamentCard
                key={tournament.id}
                tournament={tournament}
                leagueId={leagueId}
              />
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
