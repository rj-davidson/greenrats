import { z } from "zod";

export const leagueRoleSchema = z.enum(["owner", "member"]);

export type LeagueRole = z.infer<typeof leagueRoleSchema>;

export const leagueSchema = z.object({
  id: z.string(),
  name: z.string(),
  code: z.string(),
  season_year: z.number(),
  created_at: z.string(),
  role: leagueRoleSchema.optional(),
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
