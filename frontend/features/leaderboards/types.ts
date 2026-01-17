import { z } from "zod";

export const leaderboardEntrySchema = z.object({
  userId: z.string(),
  userName: z.string(),
  earnings: z.number(),
  rank: z.number(),
});

export type LeaderboardEntry = z.infer<typeof leaderboardEntrySchema>;
