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

export const fieldEntrySchema = z.object({
  id: z.string().uuid(),
  golfer_id: z.string().uuid(),
  golfer_name: z.string(),
  country_code: z.string(),
  entry_status: z.enum(["confirmed", "alternate", "withdrawn", "pending"]),
  qualifier: z.string().nullable(),
  owgr_at_entry: z.number().nullable(),
  is_amateur: z.boolean(),
});

export type FieldEntry = z.infer<typeof fieldEntrySchema>;

export const listFieldResponseSchema = z.object({
  entries: z.array(fieldEntrySchema),
  total: z.number(),
});

export type ListFieldResponse = z.infer<typeof listFieldResponseSchema>;

export const addFieldEntryRequestSchema = z.object({
  golfer_id: z.string().uuid(),
  entry_status: z.string().optional(),
  qualifier: z.string().optional().nullable(),
});

export type AddFieldEntryRequest = z.infer<typeof addFieldEntryRequestSchema>;

export const updateFieldEntryRequestSchema = z.object({
  entry_status: z.string().optional(),
  qualifier: z.string().optional().nullable(),
  owgr_at_entry: z.number().optional().nullable(),
  is_amateur: z.boolean().optional(),
});

export type UpdateFieldEntryRequest = z.infer<typeof updateFieldEntryRequestSchema>;
