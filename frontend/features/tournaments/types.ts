import { z } from "zod";

export const tournamentStatusSchema = z.enum(["upcoming", "active", "completed"]);

export const tournamentSchema = z.object({
  id: z.string(),
  name: z.string(),
  start_date: z.string(),
  end_date: z.string(),
  status: tournamentStatusSchema,
  course: z.string().optional(),
  purse: z.number().optional(),
  city: z.string().optional(),
  state: z.string().optional(),
  country: z.string().optional(),
  timezone: z.string().optional(),
  pick_window_opens_at: z.string().optional(),
  pick_window_closes_at: z.string().optional(),
  champion_id: z.string().optional(),
  champion_name: z.string().optional(),
});

export const listTournamentsResponseSchema = z.object({
  tournaments: z.array(tournamentSchema),
  total: z.number(),
});

export const getTournamentResponseSchema = z.object({
  tournament: tournamentSchema,
});

export const holeScoreSchema = z.object({
  hole_number: z.number(),
  par: z.number(),
  score: z.number().nullable(),
});

export const roundScoreSchema = z.object({
  round_number: z.number(),
  score: z.number().nullable(),
  par_relative_score: z.number().nullable(),
  tee_time: z.string().optional(),
  holes: z.array(holeScoreSchema).optional(),
});

export const leaderboardEntrySchema = z.object({
  position: z.number(),
  position_display: z.string(),
  previous_position: z.number().nullable().optional(),
  position_change: z.number().nullable().optional(),
  golfer_id: z.string(),
  golfer_name: z.string(),
  country_code: z.string(),
  country: z.string().optional(),
  image_url: z.string().optional(),
  score: z.number(),
  total_strokes: z.number(),
  thru: z.number(),
  current_round: z.number(),
  status: z.string(),
  earnings: z.number(),
  rounds: z.array(roundScoreSchema),
  picked_by: z.array(z.string()).optional(),
});

export const getLeaderboardResponseSchema = z.object({
  tournament_id: z.string(),
  tournament_name: z.string(),
  current_round: z.number(),
  entries: z.array(leaderboardEntrySchema),
  total: z.number(),
});

export const fieldEntrySchema = z.object({
  golfer_id: z.string(),
  golfer_name: z.string(),
  country_code: z.string(),
  country: z.string().optional(),
  owgr: z.number().nullable().optional(),
  owgr_at_entry: z.number().nullable().optional(),
  entry_status: z.string(),
  qualifier: z.string().optional(),
  is_amateur: z.boolean(),
  image_url: z.string().optional(),
});

export const getFieldResponseSchema = z.object({
  tournament_id: z.string(),
  tournament_name: z.string(),
  entries: z.array(fieldEntrySchema),
  total: z.number(),
});

export type TournamentStatus = z.infer<typeof tournamentStatusSchema>;
export type Tournament = z.infer<typeof tournamentSchema>;
export type ListTournamentsResponse = z.infer<typeof listTournamentsResponseSchema>;
export type GetTournamentResponse = z.infer<typeof getTournamentResponseSchema>;
export type HoleScore = z.infer<typeof holeScoreSchema>;
export type RoundScore = z.infer<typeof roundScoreSchema>;
export type LeaderboardEntry = z.infer<typeof leaderboardEntrySchema>;
export type GetLeaderboardResponse = z.infer<typeof getLeaderboardResponseSchema>;
export type FieldEntry = z.infer<typeof fieldEntrySchema>;
export type GetFieldResponse = z.infer<typeof getFieldResponseSchema>;
