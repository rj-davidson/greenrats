import type {
  LeaderboardEntry,
  LeagueLeaderboardResponse,
  LeagueLeaderboardResponseRaw,
  LeagueStandingsResponse,
  LeagueStandingsResponseRaw,
  PickHistory,
  StandingsEntry,
} from "@/features/leaderboards/types";
import {
  buildPositionCounts,
  buildRankCounts,
  calculateRanks,
  formatPositionDisplay,
  formatRankDisplay,
} from "@/lib/leaderboard";
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

function transformLeaderboardResponse(
  data: LeagueLeaderboardResponseRaw,
): LeagueLeaderboardResponse {
  const withRanks = calculateRanks(data.entries);
  const rankCounts = buildRankCounts(data.entries);

  const entries: LeaderboardEntry[] = withRanks.map((entry) => ({
    ...entry,
    rank_display: formatRankDisplay(entry.rank, rankCounts),
  }));

  return {
    entries,
    total: data.total,
    season_year: data.season_year,
  };
}

export function buildGetLeagueLeaderboardQueryOptions(
  leagueId: string,
  options: LeaderboardOptions = {},
  requestor: Requestor = makeClientRequest,
) {
  const params: Record<string, string> = {};
  if (options.seasonYear) params.season_year = String(options.seasonYear);

  return queryOptions<LeagueLeaderboardResponse, Error, LeagueLeaderboardResponse>({
    queryKey: buildLeagueLeaderboardKey(leagueId, options.seasonYear),
    queryFn: async () => {
      const raw = await requestor.get<LeagueLeaderboardResponseRaw>(
        `/api/v1/leagues/${leagueId}/leaderboard`,
        {
          params: Object.keys(params).length > 0 ? params : undefined,
        },
      );
      return transformLeaderboardResponse(raw);
    },
    enabled: !!leagueId,
  });
}

interface StandingsOptions {
  seasonYear?: number;
  include?: "picks";
}

function transformStandingsResponse(data: LeagueStandingsResponseRaw): LeagueStandingsResponse {
  const withRanks = calculateRanks(data.entries.map((e) => ({ ...e, earnings: e.total_earnings })));
  const rankCounts = buildRankCounts(data.entries.map((e) => ({ earnings: e.total_earnings })));

  const entries: StandingsEntry[] = withRanks.map((entry, i) => {
    const rawEntry = data.entries[i];
    const positionCounts = rawEntry.picks ? buildPositionCounts(rawEntry.picks) : undefined;

    const enrichedPicks: PickHistory[] | undefined = rawEntry.picks?.map((pick) => ({
      ...pick,
      position_display: formatPositionDisplay(pick.position, pick.status, positionCounts),
    }));

    return {
      user_id: rawEntry.user_id,
      user_display_name: rawEntry.user_display_name,
      total_earnings: rawEntry.total_earnings,
      pick_count: rawEntry.pick_count,
      has_current_pick: rawEntry.has_current_pick,
      current_pick: rawEntry.current_pick,
      active_picks: rawEntry.active_picks,
      rank: entry.rank,
      rank_display: formatRankDisplay(entry.rank, rankCounts),
      picks: enrichedPicks,
    };
  });

  return {
    entries,
    total: data.total,
    season_year: data.season_year,
    active_tournament: data.active_tournament,
    active_tournaments: data.active_tournaments,
  };
}

export function buildGetLeagueStandingsQueryOptions(
  leagueId: string,
  options: StandingsOptions = {},
  requestor: Requestor = makeClientRequest,
) {
  const params: Record<string, string> = {};
  if (options.seasonYear) params.season_year = String(options.seasonYear);
  if (options.include) params.include = options.include;

  return queryOptions<LeagueStandingsResponse, Error, LeagueStandingsResponse>({
    queryKey: buildLeagueStandingsKey(leagueId, options.seasonYear, options.include),
    queryFn: async () => {
      const raw = await requestor.get<LeagueStandingsResponseRaw>(
        `/api/v1/leagues/${leagueId}/standings`,
        {
          params: Object.keys(params).length > 0 ? params : undefined,
        },
      );
      return transformStandingsResponse(raw);
    },
    enabled: !!leagueId,
  });
}

export function useLeagueLeaderboard(leagueId: string, options: LeaderboardOptions = {}) {
  return useQuery(buildGetLeagueLeaderboardQueryOptions(leagueId, options));
}

export function useLeagueStandings(leagueId: string, options: StandingsOptions = {}) {
  return useQuery(buildGetLeagueStandingsQueryOptions(leagueId, options));
}
