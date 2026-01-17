import { api } from "@/lib/query/api-client";
import { useQuery } from "@tanstack/react-query";

import type { GetTournamentResponse, ListTournamentsResponse, TournamentStatus } from "./types";

interface ListTournamentsParams {
  season?: number;
  status?: TournamentStatus;
  limit?: number;
  offset?: number;
}

export const tournamentKeys = {
  all: ["tournaments"] as const,
  lists: () => [...tournamentKeys.all, "list"] as const,
  list: (params: ListTournamentsParams) => [...tournamentKeys.lists(), params] as const,
  details: () => [...tournamentKeys.all, "detail"] as const,
  detail: (id: string) => [...tournamentKeys.details(), id] as const,
  active: () => [...tournamentKeys.all, "active"] as const,
};

export function useTournaments(params: ListTournamentsParams = {}) {
  return useQuery({
    queryKey: tournamentKeys.list(params),
    queryFn: () =>
      api.get<ListTournamentsResponse>("/api/v1/tournaments", {
        params: Object.fromEntries(
          Object.entries(params)
            .filter(([, v]) => v !== undefined)
            .map(([k, v]) => [k, String(v)])
        ),
      }),
  });
}

export function useTournament(id: string) {
  return useQuery({
    queryKey: tournamentKeys.detail(id),
    queryFn: () => api.get<GetTournamentResponse>(`/api/v1/tournaments/${id}`),
    enabled: !!id,
  });
}

export function useActiveTournament() {
  return useQuery({
    queryKey: tournamentKeys.active(),
    queryFn: () => api.get<GetTournamentResponse>("/api/v1/tournaments/active"),
  });
}
