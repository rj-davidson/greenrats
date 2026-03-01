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

export const golferSeasonStatsSchema = z.object({
  scoring_avg: z.number().nullable().optional(),
  driving_distance: z.number().nullable().optional(),
  driving_accuracy: z.number().nullable().optional(),
  gir_pct: z.number().nullable().optional(),
  putting_avg: z.number().nullable().optional(),
  scrambling_pct: z.number().nullable().optional(),
  top_10s: z.number().nullable().optional(),
  cuts_made: z.number().nullable().optional(),
  events_played: z.number().nullable().optional(),
  wins: z.number().nullable().optional(),
  earnings: z.number().nullable().optional(),
});

export const golferBioSchema = z.object({
  height: z.string().optional(),
  weight: z.string().optional(),
  birth_date: z.string().nullable().optional(),
  birthplace_city: z.string().optional(),
  birthplace_state: z.string().optional(),
  birthplace_country: z.string().optional(),
  turned_pro: z.number().nullable().optional(),
  school: z.string().optional(),
  residence_city: z.string().optional(),
  residence_state: z.string().optional(),
  residence_country: z.string().optional(),
});

export const pickFieldEntrySchema = z.object({
  golfer_id: z.string(),
  golfer_name: z.string(),
  country_code: z.string(),
  country: z.string().optional(),
  image_url: z.string().optional(),
  entry_status: z.string(),
  qualifier: z.string().optional(),
  owgr: z.number().nullable().optional(),
  owgr_at_entry: z.number().nullable().optional(),
  season_earnings: z.number().nullable().optional(),
  is_amateur: z.boolean(),
  is_used: z.boolean(),
  used_for_tournament_id: z.string().optional(),
  used_for_tournament_name: z.string().optional(),
  season_stats: golferSeasonStatsSchema.nullable().optional(),
  bio: golferBioSchema.nullable().optional(),
});

export const pickWindowStateSchema = z.enum(["not_open", "open", "closed"]);

export const getPickFieldResponseSchema = z.object({
  tournament_id: z.string(),
  tournament_name: z.string(),
  course: z.string().optional(),
  city: z.string().optional(),
  state: z.string().optional(),
  country: z.string().optional(),
  purse: z.string().nullable().optional(),
  start_date: z.string(),
  end_date: z.string(),
  pick_window_state: pickWindowStateSchema,
  pick_window_opens_at: z.string().nullable().optional(),
  pick_window_closes_at: z.string().nullable().optional(),
  current_pick_id: z.string().optional(),
  current_pick_golfer_id: z.string().optional(),
  entries: z.array(pickFieldEntrySchema),
  total: z.number(),
  available_count: z.number(),
});

export type GolferSeasonStats = z.infer<typeof golferSeasonStatsSchema>;
export type GolferBio = z.infer<typeof golferBioSchema>;
export type PickFieldEntry = z.infer<typeof pickFieldEntrySchema>;
export type PickWindowState = z.infer<typeof pickWindowStateSchema>;
export type GetPickFieldResponse = z.infer<typeof getPickFieldResponseSchema>;
