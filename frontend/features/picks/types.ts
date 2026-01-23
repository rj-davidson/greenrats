import { z } from "zod";

export const pickSchema = z.object({
  id: z.string(),
  user_id: z.string(),
  tournament_id: z.string(),
  golfer_id: z.string(),
  league_id: z.string(),
  season_year: z.number(),
  created_at: z.string(),
  user_name: z.string().optional(),
  tournament_name: z.string().optional(),
  golfer_name: z.string().optional(),
  golfer_position: z.number().optional(),
  golfer_earnings: z.number().optional(),
});

export type Pick = z.infer<typeof pickSchema>;

export const pickRoundScoreSchema = z.object({
  round_number: z.number(),
  score: z.number().nullable(),
  par_relative_score: z.number().nullable(),
});

export const pickLeaderboardDataSchema = z.object({
  position: z.number(),
  position_display: z.string(),
  score: z.number(),
  thru: z.number(),
  current_round: z.number(),
  cut: z.boolean(),
  status: z.string(),
  earnings: z.number(),
  rounds: z.array(pickRoundScoreSchema).optional(),
});

export const leaguePickEntrySchema = z.object({
  pick_id: z.string(),
  user_id: z.string(),
  user_display_name: z.string(),
  golfer_id: z.string(),
  golfer_name: z.string(),
  golfer_country_code: z.string(),
  golfer_image_url: z.string().optional(),
  created_at: z.string(),
  leaderboard: pickLeaderboardDataSchema.nullable().optional(),
});

export const getLeaguePicksResponseSchema = z.object({
  entries: z.array(leaguePickEntrySchema),
  total: z.number(),
  members_without_picks: z.number(),
});

export type PickRoundScore = z.infer<typeof pickRoundScoreSchema>;
export type PickLeaderboardData = z.infer<typeof pickLeaderboardDataSchema>;
export type LeaguePickEntry = z.infer<typeof leaguePickEntrySchema>;
export type GetLeaguePicksResponse = z.infer<typeof getLeaguePicksResponseSchema>;

export const createPickRequestSchema = z.object({
  tournament_id: z.string().min(1, "Tournament is required"),
  golfer_id: z.string().min(1, "Golfer is required"),
  league_id: z.string().min(1, "League is required"),
});

export type CreatePickRequest = z.infer<typeof createPickRequestSchema>;

export const createPickResponseSchema = z.object({
  pick: pickSchema,
});

export type CreatePickResponse = z.infer<typeof createPickResponseSchema>;

export const listPicksResponseSchema = z.object({
  picks: z.array(pickSchema),
  total: z.number(),
});

export type ListPicksResponse = z.infer<typeof listPicksResponseSchema>;

export const availableGolferSchema = z.object({
  id: z.string(),
  name: z.string(),
  country_code: z.string(),
  country: z.string().optional(),
  owgr: z.number().optional(),
  image_url: z.string().optional(),
  is_used: z.boolean().optional(),
  used_for_tournament_id: z.string().optional(),
  used_for_tournament: z.string().optional(),
});

export type AvailableGolfer = z.infer<typeof availableGolferSchema>;

export const availableGolfersResponseSchema = z.object({
  golfers: z.array(availableGolferSchema),
  total: z.number(),
});

export type AvailableGolfersResponse = z.infer<typeof availableGolfersResponseSchema>;

export const overridePickRequestSchema = z.object({
  golfer_id: z.string().min(1, "Golfer is required"),
});

export type OverridePickRequest = z.infer<typeof overridePickRequestSchema>;

export const overridePickResponseSchema = z.object({
  pick: pickSchema,
});

export type OverridePickResponse = z.infer<typeof overridePickResponseSchema>;

export const createPickForUserRequestSchema = z.object({
  user_id: z.string().min(1, "User is required"),
  tournament_id: z.string().min(1, "Tournament is required"),
  golfer_id: z.string().min(1, "Golfer is required"),
});

export type CreatePickForUserRequest = z.infer<typeof createPickForUserRequestSchema>;

export const createPickForUserResponseSchema = z.object({
  pick: pickSchema,
});

export type CreatePickForUserResponse = z.infer<typeof createPickForUserResponseSchema>;

export const userPublicPickSchema = z.object({
  tournament_id: z.string(),
  tournament_name: z.string(),
  tournament_start_date: z.string(),
  golfer_id: z.string(),
  golfer_name: z.string(),
  position_display: z.string().optional(),
  earnings: z.number(),
});

export const userPublicPicksResponseSchema = z.object({
  picks: z.array(userPublicPickSchema),
});

export type UserPublicPick = z.infer<typeof userPublicPickSchema>;
export type UserPublicPicksResponse = z.infer<typeof userPublicPicksResponseSchema>;
