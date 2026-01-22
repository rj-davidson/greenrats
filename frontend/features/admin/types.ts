import { z } from "zod";

export const adminUserSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  display_name: z.string().nullable(),
  is_admin: z.boolean(),
  created_at: z.string(),
});

export type AdminUser = z.infer<typeof adminUserSchema>;

export const listUsersResponseSchema = z.object({
  users: z.array(adminUserSchema),
  total: z.number(),
});

export type ListUsersResponse = z.infer<typeof listUsersResponseSchema>;

export const adminLeagueSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  season_year: z.number(),
  member_count: z.number(),
  created_at: z.string(),
});

export type AdminLeague = z.infer<typeof adminLeagueSchema>;

export const listLeaguesResponseSchema = z.object({
  leagues: z.array(adminLeagueSchema),
  total: z.number(),
});

export type ListLeaguesResponse = z.infer<typeof listLeaguesResponseSchema>;

export const adminTournamentSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  status: z.string(),
  start_date: z.string(),
  end_date: z.string(),
});

export type AdminTournament = z.infer<typeof adminTournamentSchema>;

export const listTournamentsResponseSchema = z.object({
  tournaments: z.array(adminTournamentSchema),
  total: z.number(),
});

export type ListTournamentsResponse = z.infer<typeof listTournamentsResponseSchema>;

export const triggerResponseSchema = z.object({
  message: z.string(),
});

export type TriggerResponse = z.infer<typeof triggerResponseSchema>;
