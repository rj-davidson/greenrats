import type { LeagueLeaderboardResponse } from "@/features/leaderboards/types";
import { makeClientRequest } from "@/lib/query/client-requestor";
import { QueryKey } from "@/lib/query/query-keys";
import type { Requestor } from "@/lib/query/requestor";
import { queryOptions, useQuery } from "@tanstack/react-query";

export const buildLeagueLeaderboardKey = (leagueId: string) =>
  [QueryKey.LEAGUES, "leaderboard", leagueId] as const;

export function buildGetLeagueLeaderboardQueryOptions(
  leagueId: string,
  requestor: Requestor = makeClientRequest,
) {
  return queryOptions<LeagueLeaderboardResponse>({
    queryKey: buildLeagueLeaderboardKey(leagueId),
    queryFn: () =>
      requestor.get<LeagueLeaderboardResponse>(`/api/v1/leagues/${leagueId}/leaderboard`),
    enabled: !!leagueId,
  });
}

export function useLeagueLeaderboard(leagueId: string) {
  return useQuery(buildGetLeagueLeaderboardQueryOptions(leagueId));
}
