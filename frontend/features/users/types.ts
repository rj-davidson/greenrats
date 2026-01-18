import { z } from "zod";

export const userSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  display_name: z.string().nullable(),
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

export type CheckDisplayNameResponse = z.infer<
  typeof checkDisplayNameResponseSchema
>;
