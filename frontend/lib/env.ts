import { createEnv } from "@t3-oss/env-nextjs";
import { z } from "zod";

export const env = createEnv({
  server: {
    BASE_URL: z.url().default("http://localhost:3000"),
    WORKOS_CLIENT_ID: z.string().min(1),
    WORKOS_API_KEY: z.string().min(1),
    WORKOS_REDIRECT_URI: z.url(),
    WORKOS_COOKIE_PASSWORD: z.string().min(32),
    PRIVATE_API_URL: z.url().optional(),
  },
  client: {
    NEXT_PUBLIC_API_URL: z.url().default("http://localhost:8000"),
    NEXT_PUBLIC_SENTRY_DSN: z.string().optional(),
  },
  experimental__runtimeEnv: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
    NEXT_PUBLIC_SENTRY_DSN: process.env.NEXT_PUBLIC_SENTRY_DSN,
  },
  skipValidation: !!process.env.SKIP_ENV_VALIDATION,
});
