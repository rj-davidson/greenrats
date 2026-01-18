import { z } from "zod";

export const userSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  display_name: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
});

export type User = z.infer<typeof userSchema>;
