import type {
  CheckDisplayNameResponse,
  PendingActionsResponse,
  SetDisplayNameRequest,
  User,
} from "@/features/users/types";
import { makeClientRequest } from "@/lib/query/client-requestor";
import type { Requestor } from "@/lib/query/requestor";
import { queryOptions, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

export const userKeys = {
  all: ["users"] as const,
  me: () => [...userKeys.all, "me"] as const,
  checkDisplayName: (name: string) => [...userKeys.all, "check-display-name", name] as const,
  pendingActions: () => [...userKeys.all, "pending-actions"] as const,
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

/**
 * Hook to set the current user's display name.
 * This can only be called once - display names cannot be changed after being set.
 */
export function useSetDisplayName() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: SetDisplayNameRequest) =>
      makeClientRequest.post<User>("/api/v1/users/me/display-name", data),
    onSuccess: (user) => {
      queryClient.setQueryData(userKeys.me(), user);
    },
  });
}

/**
 * Hook to check if a display name is available.
 */
export function useCheckDisplayName(name: string) {
  return useQuery({
    queryKey: userKeys.checkDisplayName(name),
    queryFn: () =>
      makeClientRequest.get<CheckDisplayNameResponse>(
        `/api/v1/users/check-display-name?name=${encodeURIComponent(name)}`,
      ),
    enabled: name.length >= 3,
    staleTime: 30 * 1000, // 30 seconds
  });
}

export function buildGetPendingActionsQueryOptions(requestor: Requestor = makeClientRequest) {
  return queryOptions<PendingActionsResponse>({
    queryKey: userKeys.pendingActions(),
    queryFn: () => requestor.get<PendingActionsResponse>("/api/v1/users/me/pending-actions"),
    staleTime: 60 * 1000, // 1 minute
  });
}

export function usePendingActions() {
  return useQuery(buildGetPendingActionsQueryOptions());
}
