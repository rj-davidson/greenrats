import { QueryClient, isServer } from "@tanstack/react-query";

function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 60 * 1000, // 1 minute
        gcTime: 5 * 60 * 1000, // 5 minutes
        refetchOnWindowFocus: false,
        retry: 1,
      },
      dehydrate: {
        // Include pending queries in dehydration for SSR streaming
        shouldDehydrateQuery: (query) =>
          query.state.status === "pending" || query.state.status === "success",
      },
    },
  });
}

let browserQueryClient: QueryClient | undefined = undefined;

/**
 * Get the query client for the current environment.
 * - Server: Creates a new client for each request
 * - Browser: Reuses a singleton client
 */
export function getQueryClient() {
  if (isServer) {
    // Server: always make a new query client
    return makeQueryClient();
  }
  // Browser: make a new query client if we don't already have one
  if (!browserQueryClient) {
    browserQueryClient = makeQueryClient();
  }
  return browserQueryClient;
}
