import { z } from "zod";

export const tournamentStatusSchema = z.enum(["upcoming", "active", "completed"]);

export const tournamentSchema = z.object({
  id: z.string(),
  name: z.string(),
  start_date: z.string(),
  end_date: z.string(),
  status: tournamentStatusSchema,
  venue: z.string().optional(),
  course: z.string().optional(),
  purse: z.number().optional(),
});

export const listTournamentsResponseSchema = z.object({
  tournaments: z.array(tournamentSchema),
  total: z.number(),
});

export const getTournamentResponseSchema = z.object({
  tournament: tournamentSchema,
});

export type TournamentStatus = z.infer<typeof tournamentStatusSchema>;
export type Tournament = z.infer<typeof tournamentSchema>;
export type ListTournamentsResponse = z.infer<typeof listTournamentsResponseSchema>;
export type GetTournamentResponse = z.infer<typeof getTournamentResponseSchema>;
