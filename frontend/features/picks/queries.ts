import type {
  AvailableGolfersResponse,
  CreatePickRequest,
  CreatePickResponse,
  ListPicksResponse,
  OverridePickResponse,
  PickWindowStatus,
} from "./types";
import { makeClientRequest } from "@/lib/query/client-requestor";
import { QueryKey } from "@/lib/query/query-keys";
import type { Requestor } from "@/lib/query/requestor";
import { queryOptions, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

export const buildUserPicksKey = (leagueId?: string, seasonYear?: number) => {
  const key: (string | number)[] = [QueryKey.PICKS, "user"];
  if (leagueId) key.push(leagueId);
  if (seasonYear) key.push(seasonYear);
  return key;
};

export const buildLeaguePicksKey = (leagueId: string, tournamentId: string) =>
  [QueryKey.PICKS, "league", leagueId, tournamentId] as const;

export const buildAvailableGolfersKey = (leagueId: string, tournamentId: string) =>
  [QueryKey.PICKS, "available-golfers", leagueId, tournamentId] as const;

export const buildPickWindowKey = (tournamentId: string) =>
  [QueryKey.PICKS, "pick-window", tournamentId] as const;

export function buildGetUserPicksQueryOptions(
  leagueId?: string,
  seasonYear?: number,
  requestor: Requestor = makeClientRequest,
) {
  const params: Record<string, string> = {};
  if (leagueId) params.league_id = leagueId;
  if (seasonYear) params.season_year = String(seasonYear);

  return queryOptions<ListPicksResponse>({
    queryKey: buildUserPicksKey(leagueId, seasonYear),
    queryFn: () =>
      requestor.get<ListPicksResponse>("/api/v1/picks", {
        params: Object.keys(params).length > 0 ? params : undefined,
      }),
  });
}

export function buildGetLeaguePicksQueryOptions(
  leagueId: string,
  tournamentId: string,
  requestor: Requestor = makeClientRequest,
) {
  return queryOptions<ListPicksResponse>({
    queryKey: buildLeaguePicksKey(leagueId, tournamentId),
    queryFn: () =>
      requestor.get<ListPicksResponse>(`/api/v1/leagues/${leagueId}/picks`, {
        params: { tournament_id: tournamentId },
      }),
    enabled: !!leagueId && !!tournamentId,
  });
}

export function buildGetAvailableGolfersQueryOptions(
  leagueId: string,
  tournamentId: string,
  requestor: Requestor = makeClientRequest,
) {
  return queryOptions<AvailableGolfersResponse>({
    queryKey: buildAvailableGolfersKey(leagueId, tournamentId),
    queryFn: () =>
      requestor.get<AvailableGolfersResponse>(`/api/v1/leagues/${leagueId}/available-golfers`, {
        params: { tournament_id: tournamentId },
      }),
    enabled: !!leagueId && !!tournamentId,
  });
}

export function buildGetPickWindowQueryOptions(
  tournamentId: string,
  requestor: Requestor = makeClientRequest,
) {
  return queryOptions<PickWindowStatus>({
    queryKey: buildPickWindowKey(tournamentId),
    queryFn: () =>
      requestor.get<PickWindowStatus>(`/api/v1/tournaments/${tournamentId}/pick-window`),
    enabled: !!tournamentId,
  });
}

export function useUserPicks(leagueId?: string, seasonYear?: number) {
  return useQuery(buildGetUserPicksQueryOptions(leagueId, seasonYear));
}

export function useLeaguePicks(leagueId: string, tournamentId: string) {
  return useQuery(buildGetLeaguePicksQueryOptions(leagueId, tournamentId));
}

export function useAvailableGolfers(leagueId: string, tournamentId: string) {
  return useQuery(buildGetAvailableGolfersQueryOptions(leagueId, tournamentId));
}

export function usePickWindow(tournamentId: string) {
  return useQuery(buildGetPickWindowQueryOptions(tournamentId));
}

export function useCreatePick() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreatePickRequest) => {
      return makeClientRequest.post<CreatePickResponse>("/api/v1/picks", data);
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: [QueryKey.PICKS, "user"] });
      queryClient.invalidateQueries({
        queryKey: buildLeaguePicksKey(variables.league_id, variables.tournament_id),
      });
      queryClient.invalidateQueries({
        queryKey: buildAvailableGolfersKey(variables.league_id, variables.tournament_id),
      });
    },
  });
}

export function useOverridePick() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      leagueId,
      pickId,
      golferId,
    }: {
      leagueId: string;
      pickId: string;
      golferId: string;
    }) => {
      return makeClientRequest.put<OverridePickResponse>(
        `/api/v1/leagues/${leagueId}/picks/${pickId}`,
        { golfer_id: golferId },
      );
    },
    onSuccess: (_data, { leagueId }) => {
      queryClient.invalidateQueries({ queryKey: [QueryKey.PICKS, "league", leagueId] });
    },
  });
}
