"use client";

import { Skeleton } from "@/components/shadcn/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/shadcn/table";
import { useUserLeagues } from "@/features/leagues/queries";
import { useUserPicks } from "@/features/picks/queries";
import { useLeaderboard } from "@/features/tournaments/queries";
import type { LeaderboardEntry } from "@/features/tournaments/types";
import { cn } from "@/lib/utils";

interface UserPicksByLeagueProps {
  tournamentId: string;
}

function formatScore(score: number): string {
  if (score === 0) return "E";
  return score > 0 ? `+${score}` : `${score}`;
}

function formatThru(thru: number, status: string): string {
  if (status === "finished") return "F";
  if (thru === 0) return "-";
  if (thru === 18) return "F";
  return `${thru}`;
}

function formatEarnings(earnings: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(earnings);
}

function PickRow({ entry, showEarnings }: { entry: LeaderboardEntry; showEarnings: boolean }) {
  return (
    <TableRow>
      <TableCell className="font-medium">{entry.position_display}</TableCell>
      <TableCell>{entry.golfer_name}</TableCell>
      <TableCell className="text-muted-foreground">{entry.country_code}</TableCell>
      <TableCell
        className={cn("font-mono", entry.score < 0 && "text-green-600 dark:text-green-400")}
      >
        {formatScore(entry.score)}
      </TableCell>
      <TableCell className="text-muted-foreground">{formatThru(entry.thru, entry.status)}</TableCell>
      {showEarnings && (
        <TableCell className="text-right font-mono">
          {entry.earnings > 0 ? formatEarnings(entry.earnings) : "-"}
        </TableCell>
      )}
    </TableRow>
  );
}

function LeaguePickSection({
  leagueName,
  entry,
  showEarnings,
}: {
  leagueName: string;
  entry: LeaderboardEntry;
  showEarnings: boolean;
}) {
  return (
    <div className="mb-6">
      <h3 className="mb-2 text-lg font-semibold">{leagueName}</h3>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-16">Pos</TableHead>
            <TableHead>Player</TableHead>
            <TableHead className="w-20">Country</TableHead>
            <TableHead className="w-20">Score</TableHead>
            <TableHead className="w-16">Thru</TableHead>
            {showEarnings && <TableHead className="w-28 text-right">Earnings</TableHead>}
          </TableRow>
        </TableHeader>
        <TableBody>
          <PickRow entry={entry} showEarnings={showEarnings} />
        </TableBody>
      </Table>
    </div>
  );
}

function UserPicksByLeagueSkeleton() {
  return (
    <div className="mb-8 space-y-4">
      <Skeleton className="h-6 w-40" />
      <Skeleton className="h-16 w-full" />
    </div>
  );
}

export function UserPicksByLeague({ tournamentId }: UserPicksByLeagueProps) {
  const { data: leaguesData, isLoading: leaguesLoading } = useUserLeagues();
  const { data: picksData, isLoading: picksLoading } = useUserPicks();
  const { data: leaderboardData, isLoading: leaderboardLoading } = useLeaderboard(tournamentId);

  const isLoading = leaguesLoading || picksLoading || leaderboardLoading;

  if (isLoading) {
    return <UserPicksByLeagueSkeleton />;
  }

  if (!leaguesData || !picksData || !leaderboardData) {
    return null;
  }

  const tournamentPicks = picksData.picks.filter((pick) => pick.tournament_id === tournamentId);

  if (tournamentPicks.length === 0) {
    return (
      <div className="mb-8 rounded-lg border border-dashed p-4 text-center">
        <p className="text-muted-foreground">You have no picks for this tournament</p>
      </div>
    );
  }

  const leagueMap = new Map(leaguesData.leagues.map((league) => [league.id, league.name]));
  const leaderboardMap = new Map(leaderboardData.entries.map((entry) => [entry.golfer_id, entry]));
  const showEarnings = leaderboardData.entries.some((entry) => entry.earnings > 0);

  const picksWithData = tournamentPicks
    .map((pick) => {
      const leagueName = leagueMap.get(pick.league_id);
      const entry = leaderboardMap.get(pick.golfer_id);
      return { pick, leagueName, entry };
    })
    .filter(
      (item): item is { pick: typeof item.pick; leagueName: string; entry: LeaderboardEntry } =>
        !!item.leagueName && !!item.entry,
    );

  if (picksWithData.length === 0) {
    return null;
  }

  return (
    <div className="mb-8">
      <h2 className="mb-4 text-xl font-semibold">Your Picks</h2>
      {picksWithData.map(({ pick, leagueName, entry }) => (
        <LeaguePickSection
          key={pick.id}
          leagueName={leagueName}
          entry={entry}
          showEarnings={showEarnings}
        />
      ))}
    </div>
  );
}
