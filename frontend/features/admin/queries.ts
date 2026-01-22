import type {
  ListLeaguesResponse,
  ListTournamentsResponse,
  ListUsersResponse,
  TriggerResponse,
} from "@/features/admin/types";
import { makeClientRequest } from "@/lib/query/client-requestor";
import type { Requestor } from "@/lib/query/requestor";
import { QueryKey } from "@/lib/query/query-keys";
import { queryOptions, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

export const adminKeys = {
  all: [QueryKey.ADMIN] as const,
  users: () => [...adminKeys.all, "users"] as const,
  leagues: () => [...adminKeys.all, "leagues"] as const,
  tournaments: () => [...adminKeys.all, "tournaments"] as const,
};

export function buildGetAdminUsersQueryOptions(requestor: Requestor = makeClientRequest) {
  return queryOptions<ListUsersResponse>({
    queryKey: adminKeys.users(),
    queryFn: () => requestor.get<ListUsersResponse>("/api/v1/admin/users"),
  });
}

export function buildGetAdminLeaguesQueryOptions(requestor: Requestor = makeClientRequest) {
  return queryOptions<ListLeaguesResponse>({
    queryKey: adminKeys.leagues(),
    queryFn: () => requestor.get<ListLeaguesResponse>("/api/v1/admin/leagues"),
  });
}

export function buildGetAdminTournamentsQueryOptions(requestor: Requestor = makeClientRequest) {
  return queryOptions<ListTournamentsResponse>({
    queryKey: adminKeys.tournaments(),
    queryFn: () => requestor.get<ListTournamentsResponse>("/api/v1/admin/tournaments"),
  });
}

export function useAdminUsers() {
  return useQuery(buildGetAdminUsersQueryOptions());
}

export function useAdminLeagues() {
  return useQuery(buildGetAdminLeaguesQueryOptions());
}

export function useAdminTournaments() {
  return useQuery(buildGetAdminTournamentsQueryOptions());
}

export function useDeleteLeague() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (leagueId: string) => {
      await makeClientRequest.del(`/api/v1/admin/leagues/${leagueId}`);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: adminKeys.leagues() });
    },
  });
}

export function useTriggerSyncTournaments() {
  return useMutation({
    mutationFn: async () => {
      return makeClientRequest.post<TriggerResponse>("/api/v1/admin/automations/sync-tournaments");
    },
  });
}

export function useTriggerSyncPlayers() {
  return useMutation({
    mutationFn: async () => {
      return makeClientRequest.post<TriggerResponse>("/api/v1/admin/automations/sync-players");
    },
  });
}

export function useTriggerSyncLeaderboard() {
  return useMutation({
    mutationFn: async (tournamentId: string) => {
      return makeClientRequest.post<TriggerResponse>(
        `/api/v1/admin/automations/sync-leaderboard/${tournamentId}`,
      );
    },
  });
}

export function useTriggerSyncEarnings() {
  return useMutation({
    mutationFn: async (tournamentId: string) => {
      return makeClientRequest.post<TriggerResponse>(
        `/api/v1/admin/automations/sync-earnings/${tournamentId}`,
      );
    },
  });
}

export function useTriggerSyncField() {
  return useMutation({
    mutationFn: async (tournamentId: string) => {
      return makeClientRequest.post<TriggerResponse>(
        `/api/v1/admin/automations/sync-field/${tournamentId}`,
      );
    },
  });
}
