import { z } from "zod";

export const leagueRoleSchema = z.enum(["owner", "member"]);

export type LeagueRole = z.infer<typeof leagueRoleSchema>;

export const leagueSchema = z.object({
  id: z.string(),
  name: z.string(),
  code: z.string(),
  season_year: z.number(),
  joining_enabled: z.boolean(),
  created_at: z.string(),
  role: leagueRoleSchema.optional(),
  member_count: z.number().optional(),
});

export type League = z.infer<typeof leagueSchema>;

export const createLeagueRequestSchema = z.object({
  name: z.string().min(1, "League name is required").max(100),
});

export type CreateLeagueRequest = z.infer<typeof createLeagueRequestSchema>;

export const createLeagueResponseSchema = z.object({
  league: leagueSchema,
});

export type CreateLeagueResponse = z.infer<typeof createLeagueResponseSchema>;

export const getLeagueResponseSchema = z.object({
  league: leagueSchema,
});

export type GetLeagueResponse = z.infer<typeof getLeagueResponseSchema>;

export const listUserLeaguesResponseSchema = z.object({
  leagues: z.array(leagueSchema),
  total: z.number(),
});

export type ListUserLeaguesResponse = z.infer<typeof listUserLeaguesResponseSchema>;

export const joinLeagueRequestSchema = z.object({
  code: z.string().min(1, "Join code is required"),
});

export type JoinLeagueRequest = z.infer<typeof joinLeagueRequestSchema>;

export const joinLeagueResponseSchema = z.object({
  league: leagueSchema,
});

export type JoinLeagueResponse = z.infer<typeof joinLeagueResponseSchema>;

export const setJoiningEnabledRequestSchema = z.object({
  enabled: z.boolean(),
});

export type SetJoiningEnabledRequest = z.infer<typeof setJoiningEnabledRequestSchema>;

export const setJoiningEnabledResponseSchema = z.object({
  league: leagueSchema,
});

export type SetJoiningEnabledResponse = z.infer<typeof setJoiningEnabledResponseSchema>;

export const regenerateCodeResponseSchema = z.object({
  league: leagueSchema,
});

export type RegenerateCodeResponse = z.infer<typeof regenerateCodeResponseSchema>;

export const commissionerActionTypeSchema = z.enum([
  "pick_change",
  "join_code_reset",
  "joining_disabled",
  "joining_enabled",
]);

export type CommissionerActionType = z.infer<typeof commissionerActionTypeSchema>;

export const commissionerActionSchema = z.object({
  id: z.string(),
  action_type: commissionerActionTypeSchema,
  description: z.string(),
  metadata: z.record(z.string(), z.unknown()).optional(),
  created_at: z.string(),
  commissioner_name: z.string().optional(),
  affected_user_name: z.string().optional(),
});

export type CommissionerAction = z.infer<typeof commissionerActionSchema>;

export const commissionerActionsResponseSchema = z.object({
  actions: z.array(commissionerActionSchema),
  total: z.number(),
});

export type CommissionerActionsResponse = z.infer<typeof commissionerActionsResponseSchema>;
