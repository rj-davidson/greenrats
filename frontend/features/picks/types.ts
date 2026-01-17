import { z } from "zod";

export const pickSchema = z.object({
  id: z.string(),
  userId: z.string(),
  tournamentId: z.string(),
  golferId: z.string(),
  createdAt: z.string(),
});

export type Pick = z.infer<typeof pickSchema>;
