import { z } from "zod";

export const golferDetailSchema = z.object({
  id: z.string(),
  name: z.string(),
  first_name: z.string().optional(),
  last_name: z.string().optional(),
  country_code: z.string(),
  country: z.string().optional(),
  owgr: z.number().nullable().optional(),
  image_url: z.string().optional(),
  active: z.boolean(),
  height: z.string().optional(),
  weight: z.string().optional(),
  birth_date: z.string().optional(),
  birthplace_city: z.string().optional(),
  birthplace_state: z.string().optional(),
  birthplace_country: z.string().optional(),
  turned_pro: z.number().nullable().optional(),
  school: z.string().optional(),
  residence_city: z.string().optional(),
  residence_state: z.string().optional(),
  residence_country: z.string().optional(),
});

export const getGolferResponseSchema = z.object({
  golfer: golferDetailSchema,
});

export type GolferDetail = z.infer<typeof golferDetailSchema>;
export type GetGolferResponse = z.infer<typeof getGolferResponseSchema>;
