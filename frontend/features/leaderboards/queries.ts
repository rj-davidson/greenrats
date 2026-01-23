import type { LeagueLeaderboardResponse, LeagueStandingsResponse } from "@/features/leaderboards/types";
import { makeClientRequest } from "@/lib/query/client-requestor";
import { QueryKey } from "@/lib/query/query-keys";
import type { Requestor } from "@/lib/query/requestor";
import { queryOptions, useQuery } from "@tanstack/react-query";

export const buildLeagueLeaderboardKey = (leagueId: string, seasonYear?: number) =>
  [QueryKey.LEAGUES, "leaderboard", leagueId, seasonYear] as const;

export const buildLeagueStandingsKey = (leagueId: string, seasonYear?: number, include?: string) =>
  [QueryKey.LEAGUES, "standings", leagueId, seasonYear, include] as const;

interface LeaderboardOptions {
  seasonYear?: number;
}

export function buildGetLeagueLeaderboardQueryOptions(
  leagueId: string,
  options: LeaderboardOptions = {},
  requestor: Requestor = makeClientRequest,
) {
  const params: Record<string, string> = {};
  if (options.seasonYear) params.season_year = String(options.seasonYear);

  return queryOptions<LeagueLeaderboardResponse>({
    queryKey: buildLeagueLeaderboardKey(leagueId, options.seasonYear),
    queryFn: () =>
      requestor.get<LeagueLeaderboardResponse>(`/api/v1/leagues/${leagueId}/leaderboard`, {
        params: Object.keys(params).length > 0 ? params : undefined,
      }),
    enabled: !!leagueId,
  });
}

interface StandingsOptions {
  seasonYear?: number;
  include?: "picks";
}

export function buildGetLeagueStandingsQueryOptions(
  leagueId: string,
  options: StandingsOptions = {},
  requestor: Requestor = makeClientRequest,
) {
  const params: Record<string, string> = {};
  if (options.seasonYear) params.season_year = String(options.seasonYear);
  if (options.include) params.include = options.include;

  return queryOptions<LeagueStandingsResponse>({
    queryKey: buildLeagueStandingsKey(leagueId, options.seasonYear, options.include),
    queryFn: () =>
      requestor.get<LeagueStandingsResponse>(`/api/v1/leagues/${leagueId}/standings`, {
        params: Object.keys(params).length > 0 ? params : undefined,
      }),
    enabled: !!leagueId,
  });
}

export function useLeagueLeaderboard(leagueId: string, options: LeaderboardOptions = {}) {
  return useQuery(buildGetLeagueLeaderboardQueryOptions(leagueId, options));
}

export function useLeagueStandings(leagueId: string, options: StandingsOptions = {}) {
  return useQuery(buildGetLeagueStandingsQueryOptions(leagueId, options));
}
