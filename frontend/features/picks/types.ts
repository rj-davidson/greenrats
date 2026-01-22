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
