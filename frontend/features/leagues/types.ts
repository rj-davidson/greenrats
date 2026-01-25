import { z } from "zod";

export const leagueRoleSchema = z.enum(["owner", "member"]);

export type LeagueRole = z.infer<typeof leagueRoleSchema>;

export const recentPickSchema = z.object({
  golfer_name: z.string(),
  tournament_name: z.string(),
});

export type RecentPick = z.infer<typeof recentPickSchema>;

export const nextDeadlineSchema = z.object({
  tournament_id: z.string(),
  tournament_name: z.string(),
  deadline: z.string(),
});

export type NextDeadline = z.infer<typeof nextDeadlineSchema>;

export const leagueSchema = z.object({
  id: z.string(),
  name: z.string(),
  code: z.string(),
  season_year: z.number(),
  joining_enabled: z.boolean(),
  created_at: z.string(),
  role: leagueRoleSchema.optional(),
  member_count: z.number().optional(),
  recent_pick: recentPickSchema.optional(),
  next_deadline: nextDeadlineSchema.optional(),
});

export type League = z.infer<typeof leagueSchema>;

export const createLeagueRequestSchema = z.object({
  name: z
    .string()
    .min(1, "League name is required")
    .max(50, "League name must be 50 characters or less"),
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
  "member_removed",
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

export const leagueTournamentSchema = z.object({
  id: z.string(),
  name: z.string(),
  start_date: z.string(),
  end_date: z.string(),
  status: z.string(),
  course: z.string().optional(),
  city: z.string().optional(),
  state: z.string().optional(),
  country: z.string().optional(),
  has_user_pick: z.boolean(),
  user_pick_id: z.string().optional(),
  golfer_name: z.string().optional(),
  golfer_earnings: z.number().optional(),
  pick_count: z.number(),
  pick_window_opens_at: z.string().optional(),
  pick_window_closes_at: z.string().optional(),
});

export type LeagueTournament = z.infer<typeof leagueTournamentSchema>;

export const listLeagueTournamentsResponseSchema = z.object({
  tournaments: z.array(leagueTournamentSchema),
  total: z.number(),
});

export type ListLeagueTournamentsResponse = z.infer<typeof listLeagueTournamentsResponseSchema>;

export const memberPickSchema = z.object({
  id: z.string(),
  golfer_id: z.string(),
  golfer_name: z.string(),
});

export type MemberPick = z.infer<typeof memberPickSchema>;

export const leagueMemberSchema = z.object({
  id: z.string(),
  display_name: z.string(),
  role: leagueRoleSchema,
  joined_at: z.string(),
  pick: memberPickSchema.optional(),
});

export type LeagueMember = z.infer<typeof leagueMemberSchema>;

export const leagueMembersResponseSchema = z.object({
  members: z.array(leagueMemberSchema),
  total: z.number(),
});

export type LeagueMembersResponse = z.infer<typeof leagueMembersResponseSchema>;
