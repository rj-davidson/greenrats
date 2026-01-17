import { z } from "zod";

export const leagueSchema = z.object({
  id: z.string(),
  name: z.string(),
  createdById: z.string(),
  createdAt: z.string(),
});

export type League = z.infer<typeof leagueSchema>;
