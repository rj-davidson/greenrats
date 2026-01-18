import { z } from "zod";

export const tournamentStatusSchema = z.enum(["upcoming", "active", "completed"]);

export const tournamentSchema = z.object({
  id: z.string(),
  name: z.string(),
  start_date: z.string(),
  end_date: z.string(),
  status: tournamentStatusSchema,
  venue: z.string().optional(),
  course: z.string().optional(),
  purse: z.number().optional(),
});

export const listTournamentsResponseSchema = z.object({
  tournaments: z.array(tournamentSchema),
  total: z.number(),
});

export const getTournamentResponseSchema = z.object({
  tournament: tournamentSchema,
});

export const leaderboardEntrySchema = z.object({
  position: z.number(),
  position_display: z.string(),
  golfer_id: z.string(),
  golfer_name: z.string(),
  country_code: z.string(),
  score: z.number(),
  total_strokes: z.number(),
  thru: z.number(),
  current_round: z.number(),
  cut: z.boolean(),
  status: z.string(),
});

export const getLeaderboardResponseSchema = z.object({
  entries: z.array(leaderboardEntrySchema),
  total: z.number(),
});

export type TournamentStatus = z.infer<typeof tournamentStatusSchema>;
export type Tournament = z.infer<typeof tournamentSchema>;
export type ListTournamentsResponse = z.infer<typeof listTournamentsResponseSchema>;
export type GetTournamentResponse = z.infer<typeof getTournamentResponseSchema>;
export type LeaderboardEntry = z.infer<typeof leaderboardEntrySchema>;
export type GetLeaderboardResponse = z.infer<typeof getLeaderboardResponseSchema>;
