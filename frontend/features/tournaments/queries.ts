import type {
  GetLeaderboardResponse,
  GetTournamentResponse,
  ListTournamentsResponse,
  TournamentStatus,
} from "@/features/tournaments/types";
import { makeClientRequest } from "@/lib/query/client-requestor";
import { QueryKey } from "@/lib/query/query-keys";
import type { Requestor } from "@/lib/query/requestor";
import { queryOptions, useQuery } from "@tanstack/react-query";

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

export const buildLeaderboardKey = (id: string) =>
  [QueryKey.TOURNAMENTS, "leaderboard", id] as const;

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

export function buildGetLeaderboardQueryOptions(
  id: string,
  requestor: Requestor = makeClientRequest,
) {
  return queryOptions<GetLeaderboardResponse>({
    queryKey: buildLeaderboardKey(id),
    queryFn: () => requestor.get<GetLeaderboardResponse>(`/api/v1/tournaments/${id}/leaderboard`),
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

export function useLeaderboard(id: string) {
  return useQuery(buildGetLeaderboardQueryOptions(id));
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
