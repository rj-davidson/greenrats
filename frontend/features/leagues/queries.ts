import type {
  CreateLeagueRequest,
  CreateLeagueResponse,
  GetLeagueResponse,
  ListUserLeaguesResponse,
} from "./types";
import { makeClientRequest } from "@/lib/query/client-requestor";
import { QueryKey } from "@/lib/query/query-keys";
import type { Requestor } from "@/lib/query/requestor";
import { queryOptions, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

// Query key builders
export const buildUserLeaguesKey = () => [QueryKey.LEAGUES, "user-leagues"] as const;

export const buildLeagueDetailKey = (id: string) => [QueryKey.LEAGUES, "detail", id] as const;

// Query options builders
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

// React hooks (convenience wrappers)
export function useUserLeagues() {
  return useQuery(buildGetUserLeaguesQueryOptions());
}

export function useLeague(id: string) {
  return useQuery(buildGetLeagueQueryOptions(id));
}

// Mutations
export function useCreateLeague() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateLeagueRequest) => {
      return makeClientRequest.post<CreateLeagueResponse>("/api/v1/leagues", data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: buildUserLeaguesKey() });
    },
  });
}
