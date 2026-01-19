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
