import type {
  CommissionerActionsResponse,
  CreateLeagueRequest,
  CreateLeagueResponse,
  GetLeagueResponse,
  JoinLeagueRequest,
  JoinLeagueResponse,
  LeagueMembersResponse,
  ListLeagueTournamentsResponse,
  ListUserLeaguesResponse,
  RegenerateCodeResponse,
  SetJoiningEnabledRequest,
  SetJoiningEnabledResponse,
} from "@/features/leagues/types";
import { makeClientRequest } from "@/lib/query/client-requestor";
import { QueryKey } from "@/lib/query/query-keys";
import type { Requestor } from "@/lib/query/requestor";
import { queryOptions, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

export const buildUserLeaguesKey = () => [QueryKey.LEAGUES, "user-leagues"] as const;

export const buildLeagueDetailKey = (id: string) => [QueryKey.LEAGUES, "detail", id] as const;

export const buildCommissionerActionsKey = (leagueId: string) =>
  [QueryKey.LEAGUES, "commissioner-actions", leagueId] as const;

export const buildLeagueTournamentsKey = (leagueId: string) =>
  [QueryKey.LEAGUES, "tournaments", leagueId] as const;

export const buildLeagueMembersKey = (leagueId: string, tournamentId?: string) =>
  [QueryKey.LEAGUES, "members", leagueId, tournamentId] as const;

export function buildGetUserLeaguesQueryOptions(requestor: Requestor = makeClientRequest) {
  return queryOptions<ListUserLeaguesResponse>({
    queryKey: buildUserLeaguesKey(),
    queryFn: () => requestor.get<ListUserLeaguesResponse>("/api/v1/leagues"),
  });
}

export function buildGetLeagueQueryOptions(id: string, requestor: Requestor = makeClientRequest) {
  return queryOptions<GetLeagueResponse>({
    queryKey: buildLeagueDetailKey(id),
    queryFn: () => requestor.get<GetLeagueResponse>(`/api/v1/leagues/${id}`),
    enabled: !!id,
  });
}

export function buildGetCommissionerActionsQueryOptions(
  leagueId: string,
  requestor: Requestor = makeClientRequest,
) {
  return queryOptions<CommissionerActionsResponse>({
    queryKey: buildCommissionerActionsKey(leagueId),
    queryFn: () =>
      requestor.get<CommissionerActionsResponse>(
        `/api/v1/leagues/${leagueId}/commissioner-actions`,
      ),
    enabled: !!leagueId,
  });
}

export function buildGetLeagueTournamentsQueryOptions(
  leagueId: string,
  requestor: Requestor = makeClientRequest,
) {
  return queryOptions<ListLeagueTournamentsResponse>({
    queryKey: buildLeagueTournamentsKey(leagueId),
    queryFn: () =>
      requestor.get<ListLeagueTournamentsResponse>(`/api/v1/leagues/${leagueId}/tournaments`),
    enabled: !!leagueId,
  });
}

export function useUserLeagues() {
  return useQuery(buildGetUserLeaguesQueryOptions());
}

export function useLeague(id: string) {
  return useQuery(buildGetLeagueQueryOptions(id));
}

export function useCommissionerActions(leagueId: string) {
  return useQuery(buildGetCommissionerActionsQueryOptions(leagueId));
}

export function useLeagueTournaments(leagueId: string) {
  return useQuery(buildGetLeagueTournamentsQueryOptions(leagueId));
}

export function buildGetLeagueMembersQueryOptions(
  leagueId: string,
  tournamentId?: string,
  requestor: Requestor = makeClientRequest,
) {
  const url = tournamentId
    ? `/api/v1/leagues/${leagueId}/members?tournament_id=${tournamentId}`
    : `/api/v1/leagues/${leagueId}/members`;

  return queryOptions<LeagueMembersResponse>({
    queryKey: buildLeagueMembersKey(leagueId, tournamentId),
    queryFn: () => requestor.get<LeagueMembersResponse>(url),
    enabled: !!leagueId,
  });
}

export function useLeagueMembers(leagueId: string, tournamentId?: string) {
  return useQuery(buildGetLeagueMembersQueryOptions(leagueId, tournamentId));
}

export function useCreateLeague() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateLeagueRequest) => {
      return makeClientRequest.post<CreateLeagueResponse>("/api/v1/leagues", data);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: buildUserLeaguesKey() });
    },
  });
}

export function useJoinLeague() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: JoinLeagueRequest) => {
      return makeClientRequest.post<JoinLeagueResponse>("/api/v1/leagues/join", data);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: buildUserLeaguesKey() });
    },
  });
}

export function useRegenerateJoinCode() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (leagueId: string) => {
      return makeClientRequest.post<RegenerateCodeResponse>(
        `/api/v1/leagues/${leagueId}/regenerate-code`,
      );
    },
    onSuccess: (_data, leagueId) => {
      void queryClient.invalidateQueries({ queryKey: buildLeagueDetailKey(leagueId) });
      void queryClient.invalidateQueries({ queryKey: buildCommissionerActionsKey(leagueId) });
    },
  });
}

export function useSetJoiningEnabled() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ leagueId, enabled }: { leagueId: string; enabled: boolean }) => {
      return makeClientRequest.patch<SetJoiningEnabledResponse>(
        `/api/v1/leagues/${leagueId}/joining`,
        { enabled } as SetJoiningEnabledRequest,
      );
    },
    onSuccess: (_data, { leagueId }) => {
      void queryClient.invalidateQueries({ queryKey: buildLeagueDetailKey(leagueId) });
      void queryClient.invalidateQueries({ queryKey: buildCommissionerActionsKey(leagueId) });
    },
  });
}

export function useRemoveMember() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ leagueId, userId }: { leagueId: string; userId: string }) => {
      return makeClientRequest.del(`/api/v1/leagues/${leagueId}/members/${userId}`);
    },
    onSuccess: (_data, { leagueId }) => {
      void queryClient.invalidateQueries({ queryKey: buildLeagueDetailKey(leagueId) });
      void queryClient.invalidateQueries({ queryKey: buildLeagueMembersKey(leagueId) });
      void queryClient.invalidateQueries({ queryKey: buildCommissionerActionsKey(leagueId) });
    },
  });
}
