import { useQuery } from "@tanstack/react-query";

import { makeClientRequest } from "@/lib/query/client-requestor";
import type { User } from "./types";

export const userKeys = {
  all: ["users"] as const,
  me: () => [...userKeys.all, "me"] as const,
};

/**
 * Hook to fetch the current authenticated user from the database.
 * This returns the DB user (with display_name, etc.) rather than the WorkOS user.
 */
export function useCurrentUser() {
  return useQuery({
    queryKey: userKeys.me(),
    queryFn: () => makeClientRequest.get<User>("/api/v1/users/me"),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}
