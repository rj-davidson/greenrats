"use client";

import { useBreadcrumbs } from "@/components/core/breadcrumbs";
import { Badge } from "@/components/shadcn/badge";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/shadcn/empty";
import { Skeleton } from "@/components/shadcn/skeleton";
import { useLeague } from "@/features/leagues/queries";
import {
  PickFieldTable,
  PickFieldSkeleton,
  TournamentPickHeader,
} from "@/features/picks/components";
import { useLeaguePicks, usePickField } from "@/features/picks/queries";
import { formatCountdown, formatPickWindowDate } from "@/features/picks/utils";
import { LiveExpandableLeaderboard } from "@/features/tournaments/components/live";
import { PlacementExpandableLeaderboard } from "@/features/tournaments/components/placement";
import { useTournament } from "@/features/tournaments/queries";
import { useCurrentUser } from "@/features/users/queries";
import { CalendarIcon, ClockIcon } from "lucide-react";
import { useParams } from "next/navigation";
import { useEffect, useMemo } from "react";

function formatDateRange(startDate: string, endDate: string): string {
  const start = new Date(startDate);
  const end = new Date(endDate);
  const options: Intl.DateTimeFormatOptions = { month: "short", day: "numeric", year: "numeric" };

  return `${start.toLocaleDateString("en-US", options)} - ${end.toLocaleDateString("en-US", options)}`;
}

export default function TournamentDetailPage() {
  const params = useParams<{ leagueId: string; tournamentId: string }>();
  const { leagueId, tournamentId } = params;

  const { data: leagueData } = useLeague(leagueId);
  const { data: tournamentData, isLoading: tournamentLoading } = useTournament(tournamentId);
  const { data: picksData } = useLeaguePicks(leagueId, tournamentId);
  const { data: pickFieldData, isLoading: pickFieldLoading } = usePickField(leagueId, tournamentId);
  const { data: currentUser } = useCurrentUser();
  const { setExtraCrumbs } = useBreadcrumbs();

  const league = leagueData?.league;
  const tournament = tournamentData?.tournament;

  const userPickedGolferId = useMemo(() => {
    if (!currentUser || !picksData?.entries) return undefined;
    const entry = picksData.entries.find((p) => p.user_id === currentUser.id);
    return entry?.golfer_id;
  }, [currentUser, picksData]);

  const currentPickGolferName = useMemo(() => {
    if (!pickFieldData?.current_pick_golfer_id) return undefined;
    const entry = pickFieldData.entries.find((e) => e.golfer_id === pickFieldData.current_pick_golfer_id);
    return entry?.golfer_name;
  }, [pickFieldData]);

  useEffect(() => {
    const crumbs: { name: string; path?: string }[] = [];
    if (league?.name) {
      crumbs.push({ name: league.name, path: `/${leagueId}` });
    }
    if (tournament?.name) {
      crumbs.push({ name: tournament.name });
    }
    setExtraCrumbs(crumbs);
    return () => setExtraCrumbs([]);
  }, [league?.name, tournament?.name, leagueId, setExtraCrumbs]);

  if (tournamentLoading || pickFieldLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-20 w-full" />
        <PickFieldSkeleton />
      </div>
    );
  }

  if (!tournament) {
    return (
      <Empty>
        <EmptyHeader>
          <EmptyTitle>Tournament Not Found</EmptyTitle>
          <EmptyDescription>The tournament you are looking for does not exist.</EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  if (pickFieldData?.pick_window_state === "open") {
    return (
      <div className="space-y-6">
        <TournamentPickHeader data={pickFieldData} currentPickGolferName={currentPickGolferName} />
        <PickFieldTable data={pickFieldData} leagueId={leagueId} />
      </div>
    );
  }

  if (pickFieldData?.pick_window_state === "not_open") {
    const opensAt = pickFieldData.pick_window_opens_at;
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold">{tournament.name}</h1>
          <div className="mt-1 flex items-center gap-2 text-muted-foreground">
            <CalendarIcon className="size-4" />
            {formatDateRange(tournament.start_date, tournament.end_date)}
          </div>
        </div>
        <Empty>
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <ClockIcon />
            </EmptyMedia>
            <EmptyTitle>Pick window not yet open</EmptyTitle>
            <EmptyDescription>
              Check back soon to make your pick for this tournament.
            </EmptyDescription>
          </EmptyHeader>
          {opensAt && (
            <EmptyContent>
              <div className="text-center">
                <p className="text-sm text-muted-foreground">Opens in</p>
                <p className="text-2xl font-bold">{formatCountdown(opensAt)}</p>
                <p className="text-xs text-muted-foreground">{formatPickWindowDate(opensAt)}</p>
              </div>
            </EmptyContent>
          )}
        </Empty>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-bold">{tournament.name}</h1>
          {tournament.status === "active" && (
            <Badge variant="default" className="text-xs">
              Live
            </Badge>
          )}
        </div>
        <div className="mt-1 flex items-center gap-2 text-muted-foreground">
          <CalendarIcon className="size-4" />
          {formatDateRange(tournament.start_date, tournament.end_date)}
        </div>
      </div>
      {tournament.status === "completed" ? (
        <PlacementExpandableLeaderboard
          tournamentId={tournamentId}
          leagueId={leagueId}
          highlightedGolferId={userPickedGolferId}
        />
      ) : (
        <LiveExpandableLeaderboard
          tournamentId={tournamentId}
          leagueId={leagueId}
          highlightedGolferId={userPickedGolferId}
        />
      )}
    </div>
  );
}
