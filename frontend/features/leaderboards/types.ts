import { z } from "zod";

export const leaderboardEntrySchema = z.object({
  rank: z.number(),
  user_id: z.string(),
  display_name: z.string(),
  earnings: z.number(),
  pick_count: z.number(),
});

export type LeaderboardEntry = z.infer<typeof leaderboardEntrySchema>;

export const leagueLeaderboardResponseSchema = z.object({
  entries: z.array(leaderboardEntrySchema),
  total: z.number(),
  season_year: z.number(),
});

export type LeagueLeaderboardResponse = z.infer<typeof leagueLeaderboardResponseSchema>;

export const pickHistorySchema = z.object({
  tournament_id: z.string(),
  tournament_name: z.string(),
  golfer_id: z.string(),
  golfer_name: z.string(),
  position_display: z.string().optional(),
  earnings: z.number(),
});

export const standingsEntrySchema = z.object({
  rank: z.number(),
  user_id: z.string(),
  user_display_name: z.string(),
  total_earnings: z.number(),
  pick_count: z.number(),
  picks: z.array(pickHistorySchema).optional(),
});

export const leagueStandingsResponseSchema = z.object({
  entries: z.array(standingsEntrySchema),
  total: z.number(),
  season_year: z.number(),
});

export type PickHistory = z.infer<typeof pickHistorySchema>;
export type StandingsEntry = z.infer<typeof standingsEntrySchema>;
export type LeagueStandingsResponse = z.infer<typeof leagueStandingsResponseSchema>;
