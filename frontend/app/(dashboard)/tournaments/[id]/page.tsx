import { Skeleton } from "@/components/shadcn/skeleton";
import { buildGetTournamentQueryOptions } from "@/features/tournaments/queries";
import { makeServerRequest } from "@/lib/query/server-requestor";
import { HydrationBoundary, QueryClient, dehydrate } from "@tanstack/react-query";
import { TournamentCoverage } from "./tournament-coverage";

interface TournamentPageProps {
  params: Promise<{ id: string }>;
}

export default async function TournamentPage({ params }: TournamentPageProps) {
  const { id } = await params;

  const queryClient = new QueryClient();
  await queryClient.prefetchQuery(buildGetTournamentQueryOptions(id, makeServerRequest));

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <TournamentCoverage id={id} />
    </HydrationBoundary>
  );
}

export function TournamentPageSkeleton() {
  return (
    <div className="container mx-auto p-8">
      <Skeleton className="mb-2 h-9 w-64" />
      <Skeleton className="h-5 w-96" />
    </div>
  );
}
