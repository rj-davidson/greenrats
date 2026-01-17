import { z } from "zod";

export const tournamentSchema = z.object({
  id: z.string(),
  name: z.string(),
  startDate: z.string(),
  endDate: z.string(),
  status: z.enum(["upcoming", "active", "completed"]),
});

export type Tournament = z.infer<typeof tournamentSchema>;
