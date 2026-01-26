import type {
  GetFieldResponse,
  GetLeaderboardResponse,
  GetLeaderboardResponseRaw,
  GetTournamentResponse,
  LeaderboardEntry,
  ListTournamentsResponse,
  TournamentStatus,
} from "@/features/tournaments/types";
import { buildPositionCounts, formatPositionDisplay } from "@/lib/leaderboard";
import { makeClientRequest } from "@/lib/query/client-requestor";
import { QueryKey } from "@/lib/query/query-keys";
import type { Requestor } from "@/lib/query/requestor";
import { queryOptions, useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback } from "react";

interface ListTournamentsParams {
  season?: number;
  status?: TournamentStatus;
  limit?: number;
  offset?: number;
}

// Query key builders
export const buildTournamentListKey = (params: ListTournamentsParams = {}) =>
  [QueryKey.TOURNAMENTS, "list", params] as const;

export const buildTournamentDetailKey = (id: string) =>
  [QueryKey.TOURNAMENTS, "detail", id] as const;

export const buildTournamentActiveKey = () => [QueryKey.TOURNAMENTS, "active"] as const;

export const buildLeaderboardKey = (id: string, include?: string, leagueId?: string) =>
  [QueryKey.TOURNAMENTS, "leaderboard", id, include, leagueId] as const;

export const buildFieldKey = (id: string) => [QueryKey.TOURNAMENTS, "field", id] as const;

// Query options builders
export function buildGetTournamentsQueryOptions(
  params: ListTournamentsParams = {},
  requestor: Requestor = makeClientRequest,
) {
  const queryParams = Object.fromEntries(
    Object.entries(params)
      .filter(([, v]) => v !== undefined)
      .map(([k, v]) => [k, String(v)]),
  );

  return queryOptions<ListTournamentsResponse>({
    queryKey: buildTournamentListKey(params),
    queryFn: () =>
      requestor.get<ListTournamentsResponse>("/api/v1/tournaments", {
        params: Object.keys(queryParams).length > 0 ? queryParams : undefined,
      }),
  });
}

export function buildGetTournamentQueryOptions(
  id: string,
  requestor: Requestor = makeClientRequest,
) {
  return queryOptions<GetTournamentResponse>({
    queryKey: buildTournamentDetailKey(id),
    queryFn: () => requestor.get<GetTournamentResponse>(`/api/v1/tournaments/${id}`),
    enabled: !!id,
  });
}

export function buildGetActiveTournamentQueryOptions(requestor: Requestor = makeClientRequest) {
  return queryOptions<GetTournamentResponse>({
    queryKey: buildTournamentActiveKey(),
    queryFn: () => requestor.get<GetTournamentResponse>("/api/v1/tournaments/active"),
  });
}

interface LeaderboardOptions {
  include?: "holes";
  leagueId?: string;
}

function transformLeaderboardResponse(data: GetLeaderboardResponseRaw): GetLeaderboardResponse {
  const positionCounts = buildPositionCounts(data.entries);

  const entries: LeaderboardEntry[] = data.entries.map((entry) => ({
    ...entry,
    position_display: formatPositionDisplay(entry.position, entry.status, positionCounts),
  }));

  return {
    tournament_id: data.tournament_id,
    tournament_name: data.tournament_name,
    current_round: data.current_round,
    entries,
    total: data.total,
  };
}

export function buildGetLeaderboardQueryOptions(
  id: string,
  options: LeaderboardOptions = {},
  requestor: Requestor = makeClientRequest,
) {
  const params: Record<string, string> = {};
  if (options.include) params.include = options.include;
  if (options.leagueId) params.league_id = options.leagueId;

  return queryOptions<GetLeaderboardResponse, Error, GetLeaderboardResponse>({
    queryKey: buildLeaderboardKey(id, options.include, options.leagueId),
    queryFn: async () => {
      const raw = await requestor.get<GetLeaderboardResponseRaw>(
        `/api/v1/tournaments/${id}/leaderboard`,
        {
          params: Object.keys(params).length > 0 ? params : undefined,
        },
      );
      return transformLeaderboardResponse(raw);
    },
    enabled: !!id,
  });
}

export function buildGetFieldQueryOptions(id: string, requestor: Requestor = makeClientRequest) {
  return queryOptions<GetFieldResponse>({
    queryKey: buildFieldKey(id),
    queryFn: () => requestor.get<GetFieldResponse>(`/api/v1/tournaments/${id}/field`),
    enabled: !!id,
  });
}

// React hooks (convenience wrappers)
export function useTournaments(params: ListTournamentsParams = {}) {
  return useQuery(buildGetTournamentsQueryOptions(params));
}

export function useTournament(id: string) {
  return useQuery(buildGetTournamentQueryOptions(id));
}

export function useActiveTournament() {
  return useQuery(buildGetActiveTournamentQueryOptions());
}

export function useLeaderboard(id: string, options: LeaderboardOptions = {}) {
  return useQuery(buildGetLeaderboardQueryOptions(id, options));
}

export function useTournamentField(id: string) {
  return useQuery(buildGetFieldQueryOptions(id));
}

export function usePrefetchLeaderboardWithHoles(tournamentId: string, leagueId?: string) {
  const queryClient = useQueryClient();
  return useCallback(() => {
    void queryClient.prefetchQuery(
      buildGetLeaderboardQueryOptions(tournamentId, { include: "holes", leagueId }),
    );
  }, [queryClient, tournamentId, leagueId]);
}

export function useCurrentTournament() {
  const { data: activeData, isLoading: activeLoading } = useActiveTournament();
  const { data: completedData, isLoading: completedLoading } = useTournaments({
    status: "completed",
    limit: 2,
  });
  const { data: upcomingData, isLoading: upcomingLoading } = useTournaments({
    status: "upcoming",
    limit: 1,
  });

  const isLoading = activeLoading || completedLoading || upcomingLoading;

  const tournament = (() => {
    if (activeData?.tournament) {
      return activeData.tournament;
    }

    const now = new Date();
    const recentCompleted = completedData?.tournaments.find((t) => {
      const endPlusOne = new Date(t.end_date);
      endPlusOne.setDate(endPlusOne.getDate() + 1);
      endPlusOne.setHours(23, 59, 59, 999);
      return now <= endPlusOne;
    });

    if (recentCompleted) {
      return recentCompleted;
    }

    return upcomingData?.tournaments[0] ?? null;
  })();

  return { tournament, isLoading };
}
