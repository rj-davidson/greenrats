import { buildLeagueTournamentsKey } from "@/features/leagues/queries";
import type {
  AvailableGolfersResponse,
  CreatePickForUserResponse,
  CreatePickRequest,
  CreatePickResponse,
  GetLeaguePicksResponse,
  GetPickFieldResponse,
  ListPicksResponse,
  OverridePickResponse,
  UserPublicPicksResponse,
} from "@/features/picks/types";
import { makeClientRequest } from "@/lib/query/client-requestor";
import { QueryKey } from "@/lib/query/query-keys";
import type { Requestor } from "@/lib/query/requestor";
import { queryOptions, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback } from "react";

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

export const buildAvailableGolfersForUserKey = (
  leagueId: string,
  tournamentId: string,
  userId: string,
) => [QueryKey.PICKS, "available-golfers-for-user", leagueId, tournamentId, userId] as const;

export const buildPickFieldKey = (leagueId: string, tournamentId: string) =>
  [QueryKey.PICKS, "pick-field", leagueId, tournamentId] as const;

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

interface LeaguePicksOptions {
  include?: "rounds";
}

export function buildGetLeaguePicksQueryOptions(
  leagueId: string,
  tournamentId: string,
  optionsOrRequestor: LeaguePicksOptions | Requestor = {},
  maybeRequestor?: Requestor,
) {
  const isRequestor = (val: unknown): val is Requestor =>
    typeof val === "object" && val !== null && "get" in val;

  const options: LeaguePicksOptions = isRequestor(optionsOrRequestor) ? {} : optionsOrRequestor;
  const requestor: Requestor = isRequestor(optionsOrRequestor)
    ? optionsOrRequestor
    : (maybeRequestor ?? makeClientRequest);

  const params: Record<string, string> = { tournament_id: tournamentId };
  if (options.include) params.include = options.include;

  return queryOptions<GetLeaguePicksResponse>({
    queryKey: buildLeaguePicksKey(leagueId, tournamentId),
    queryFn: () =>
      requestor.get<GetLeaguePicksResponse>(`/api/v1/leagues/${leagueId}/picks`, {
        params,
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

export function useUserPicks(leagueId?: string, seasonYear?: number) {
  return useQuery(buildGetUserPicksQueryOptions(leagueId, seasonYear));
}

export function useLeaguePicks(
  leagueId: string,
  tournamentId: string,
  options: LeaguePicksOptions = {},
) {
  return useQuery(buildGetLeaguePicksQueryOptions(leagueId, tournamentId, options));
}

export function useAvailableGolfers(leagueId: string, tournamentId: string) {
  return useQuery(buildGetAvailableGolfersQueryOptions(leagueId, tournamentId));
}

export function buildGetAvailableGolfersForUserQueryOptions(
  leagueId: string,
  tournamentId: string,
  userId: string,
  requestor: Requestor = makeClientRequest,
) {
  return queryOptions<AvailableGolfersResponse>({
    queryKey: buildAvailableGolfersForUserKey(leagueId, tournamentId, userId),
    queryFn: () =>
      requestor.get<AvailableGolfersResponse>(
        `/api/v1/leagues/${leagueId}/available-golfers-for-user`,
        {
          params: { tournament_id: tournamentId, user_id: userId },
        },
      ),
    enabled: !!leagueId && !!tournamentId && !!userId,
  });
}

export function useAvailableGolfersForUser(leagueId: string, tournamentId: string, userId: string) {
  return useQuery(buildGetAvailableGolfersForUserQueryOptions(leagueId, tournamentId, userId));
}

export function buildGetPickFieldQueryOptions(
  leagueId: string,
  tournamentId: string,
  requestor: Requestor = makeClientRequest,
) {
  return queryOptions<GetPickFieldResponse>({
    queryKey: buildPickFieldKey(leagueId, tournamentId),
    queryFn: () =>
      requestor.get<GetPickFieldResponse>(
        `/api/v1/leagues/${leagueId}/tournaments/${tournamentId}/pick-field`,
      ),
    enabled: !!leagueId && !!tournamentId,
  });
}

export function usePickField(leagueId: string, tournamentId: string) {
  return useQuery(buildGetPickFieldQueryOptions(leagueId, tournamentId));
}

export function useCreatePick() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreatePickRequest) => {
      return makeClientRequest.post<CreatePickResponse>("/api/v1/picks", data);
    },
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({ queryKey: [QueryKey.PICKS, "user"] });
      void queryClient.invalidateQueries({
        queryKey: buildLeaguePicksKey(variables.league_id, variables.tournament_id),
      });
      void queryClient.invalidateQueries({
        queryKey: buildAvailableGolfersKey(variables.league_id, variables.tournament_id),
      });
      void queryClient.invalidateQueries({
        queryKey: buildPickFieldKey(variables.league_id, variables.tournament_id),
      });
      void queryClient.invalidateQueries({
        queryKey: buildLeagueTournamentsKey(variables.league_id),
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
      tournamentId?: string;
    }) => {
      return makeClientRequest.put<OverridePickResponse>(
        `/api/v1/leagues/${leagueId}/picks/${pickId}`,
        { golfer_id: golferId },
      );
    },
    onSuccess: (_data, { leagueId, tournamentId }) => {
      void queryClient.invalidateQueries({ queryKey: [QueryKey.PICKS, "league", leagueId] });
      void queryClient.invalidateQueries({
        queryKey: [QueryKey.LEAGUES, "commissioner-actions", leagueId],
      });
      void queryClient.invalidateQueries({
        queryKey: buildLeagueTournamentsKey(leagueId),
      });
      if (tournamentId) {
        void queryClient.invalidateQueries({
          queryKey: [QueryKey.LEAGUES, "members", leagueId, tournamentId],
        });
      }
    },
  });
}

export function useCreatePickForUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      leagueId,
      userId,
      tournamentId,
      golferId,
    }: {
      leagueId: string;
      userId: string;
      tournamentId: string;
      golferId: string;
    }) => {
      return makeClientRequest.post<CreatePickForUserResponse>(
        `/api/v1/leagues/${leagueId}/picks/create-for-user`,
        { user_id: userId, tournament_id: tournamentId, golfer_id: golferId },
      );
    },
    onSuccess: (_data, { leagueId, tournamentId }) => {
      void queryClient.invalidateQueries({ queryKey: [QueryKey.PICKS, "league", leagueId] });
      void queryClient.invalidateQueries({
        queryKey: [QueryKey.LEAGUES, "commissioner-actions", leagueId],
      });
      void queryClient.invalidateQueries({
        queryKey: buildLeagueTournamentsKey(leagueId),
      });
      void queryClient.invalidateQueries({
        queryKey: [QueryKey.LEAGUES, "members", leagueId, tournamentId],
      });
    },
  });
}

export function useUpdatePick() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      pickId,
      golferId,
    }: {
      pickId: string;
      golferId: string;
      leagueId: string;
      tournamentId: string;
    }) => {
      return makeClientRequest.put<CreatePickResponse>(`/api/v1/picks/${pickId}`, {
        golfer_id: golferId,
      });
    },
    onSuccess: (_data, { leagueId, tournamentId }) => {
      void queryClient.invalidateQueries({ queryKey: [QueryKey.PICKS, "user"] });
      void queryClient.invalidateQueries({
        queryKey: buildLeaguePicksKey(leagueId, tournamentId),
      });
      void queryClient.invalidateQueries({
        queryKey: buildAvailableGolfersKey(leagueId, tournamentId),
      });
      void queryClient.invalidateQueries({
        queryKey: buildPickFieldKey(leagueId, tournamentId),
      });
      void queryClient.invalidateQueries({
        queryKey: buildLeagueTournamentsKey(leagueId),
      });
    },
  });
}

export const buildUserPublicPicksKey = (leagueId: string, userId: string) =>
  [QueryKey.PICKS, "public", leagueId, userId] as const;

export function buildGetUserPublicPicksQueryOptions(
  leagueId: string,
  userId: string,
  requestor: Requestor = makeClientRequest,
) {
  return queryOptions<UserPublicPicksResponse>({
    queryKey: buildUserPublicPicksKey(leagueId, userId),
    queryFn: () =>
      requestor.get<UserPublicPicksResponse>(`/api/v1/leagues/${leagueId}/users/${userId}/picks`),
    enabled: !!leagueId && !!userId,
  });
}

export function useUserPublicPicks(leagueId: string, userId: string) {
  return useQuery(buildGetUserPublicPicksQueryOptions(leagueId, userId));
}

export function usePrefetchUserPublicPicks(leagueId: string) {
  const queryClient = useQueryClient();

  return useCallback(
    (userId: string) => {
      void queryClient.prefetchQuery(buildGetUserPublicPicksQueryOptions(leagueId, userId));
    },
    [queryClient, leagueId],
  );
}
