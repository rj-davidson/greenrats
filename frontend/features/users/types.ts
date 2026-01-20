import { z } from "zod";

export const userSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  display_name: z.string().nullable(),
  is_admin: z.boolean().optional(),
  created_at: z.string(),
  updated_at: z.string(),
});

export type User = z.infer<typeof userSchema>;

export const setDisplayNameRequestSchema = z.object({
  display_name: z
    .string()
    .min(3, "Must be at least 3 characters")
    .max(20, "Must be at most 20 characters")
    .regex(/^[a-zA-Z0-9_]+$/, "Only letters, numbers, and underscores allowed"),
});

export type SetDisplayNameRequest = z.infer<typeof setDisplayNameRequestSchema>;

export const checkDisplayNameResponseSchema = z.object({
  available: z.boolean(),
  name: z.string(),
});

export type CheckDisplayNameResponse = z.infer<typeof checkDisplayNameResponseSchema>;

export const pendingPickActionSchema = z.object({
  league_id: z.string(),
  league_name: z.string(),
  tournament_id: z.string(),
  tournament_name: z.string(),
  pick_deadline: z.string(),
});

export type PendingPickAction = z.infer<typeof pendingPickActionSchema>;

export const upcomingTournamentSchema = z.object({
  id: z.string(),
  name: z.string(),
  start_date: z.string(),
  end_date: z.string(),
  status: z.string(),
});

export type UpcomingTournament = z.infer<typeof upcomingTournamentSchema>;

export const pendingActionsResponseSchema = z.object({
  pending_picks: z.array(pendingPickActionSchema),
  upcoming_tournaments: z.array(upcomingTournamentSchema),
});

export type PendingActionsResponse = z.infer<typeof pendingActionsResponseSchema>;
