/**
 * Centralized environment variables for the application.
 *
 * PUBLIC_BACKEND_URL: Used by the browser for API calls (must be NEXT_PUBLIC_*)
 * PRIVATE_BACKEND_URL: Used by server components (can be internal Docker network address)
 */

export const PUBLIC_BACKEND_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";

export const PRIVATE_BACKEND_URL =
  process.env.PRIVATE_API_URL || PUBLIC_BACKEND_URL;
