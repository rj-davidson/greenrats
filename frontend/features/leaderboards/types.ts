import { z } from "zod";

export const currentPickSchema = z.object({
  tournament_id: z.string(),
  tournament_name: z.string(),
  golfer_id: z.string(),
  golfer_name: z.string(),
});

export const activeTournamentSchema = z.object({
  id: z.string(),
  name: z.string(),
  is_pick_window_closed: z.boolean(),
  start_date: z.string(),
});

export type CurrentPick = z.infer<typeof currentPickSchema>;
export type ActiveTournament = z.infer<typeof activeTournamentSchema>;

export const activePickEntrySchema = z.object({
  tournament_id: z.string(),
  tournament_name: z.string(),
  has_pick: z.boolean(),
  golfer_id: z.string().optional(),
  golfer_name: z.string().optional(),
  is_pick_window_closed: z.boolean(),
});

export type ActivePickEntry = z.infer<typeof activePickEntrySchema>;

export const leaderboardEntryRawSchema = z.object({
  user_id: z.string(),
  display_name: z.string(),
  earnings: z.number(),
  pick_count: z.number(),
});

export type LeaderboardEntryRaw = z.infer<typeof leaderboardEntryRawSchema>;

export interface LeaderboardEntry extends LeaderboardEntryRaw {
  rank: number;
  rank_display: string;
}

export const leagueLeaderboardResponseRawSchema = z.object({
  entries: z.array(leaderboardEntryRawSchema),
  total: z.number(),
  season_year: z.number(),
});

export type LeagueLeaderboardResponseRaw = z.infer<typeof leagueLeaderboardResponseRawSchema>;

export interface LeagueLeaderboardResponse {
  entries: LeaderboardEntry[];
  total: number;
  season_year: number;
}

export const pickHistoryRawSchema = z.object({
  tournament_id: z.string(),
  tournament_name: z.string(),
  golfer_id: z.string(),
  golfer_name: z.string(),
  position: z.number(),
  status: z.string(),
  earnings: z.number(),
});

export type PickHistoryRaw = z.infer<typeof pickHistoryRawSchema>;

export interface PickHistory extends PickHistoryRaw {
  position_display: string;
}

export const standingsEntryRawSchema = z.object({
  user_id: z.string(),
  user_display_name: z.string(),
  total_earnings: z.number(),
  pick_count: z.number(),
  has_current_pick: z.boolean(),
  current_pick: currentPickSchema.optional(),
  active_picks: z.array(activePickEntrySchema).optional(),
  picks: z.array(pickHistoryRawSchema).optional(),
});

export type StandingsEntryRaw = z.infer<typeof standingsEntryRawSchema>;

export interface StandingsEntry extends Omit<StandingsEntryRaw, "picks"> {
  rank: number;
  rank_display: string;
  picks?: PickHistory[];
  active_picks?: ActivePickEntry[];
}

export const leagueStandingsResponseRawSchema = z.object({
  entries: z.array(standingsEntryRawSchema),
  total: z.number(),
  season_year: z.number(),
  active_tournament: activeTournamentSchema.optional(),
  active_tournaments: z.array(activeTournamentSchema).optional(),
});

export type LeagueStandingsResponseRaw = z.infer<typeof leagueStandingsResponseRawSchema>;

export interface LeagueStandingsResponse {
  entries: StandingsEntry[];
  total: number;
  season_year: number;
  active_tournament?: ActiveTournament;
  active_tournaments?: ActiveTournament[];
}
