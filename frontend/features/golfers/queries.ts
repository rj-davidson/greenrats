import type { GetGolferResponse } from "@/features/golfers/types";
import { makeClientRequest } from "@/lib/query/client-requestor";
import { QueryKey } from "@/lib/query/query-keys";
import type { Requestor } from "@/lib/query/requestor";
import { queryOptions, useQuery } from "@tanstack/react-query";

export const buildGolferDetailKey = (id: string) =>
  [QueryKey.GOLFERS, "detail", id] as const;

export function buildGetGolferQueryOptions(
  id: string,
  requestor: Requestor = makeClientRequest,
) {
  return queryOptions<GetGolferResponse>({
    queryKey: buildGolferDetailKey(id),
    queryFn: () => requestor.get<GetGolferResponse>(`/api/v1/golfers/${id}`),
    enabled: !!id,
  });
}

export function useGolfer(id: string) {
  return useQuery(buildGetGolferQueryOptions(id));
}
