import { LeaguePickerContent } from "@/app/league-picker";
import { TopBar } from "@/components/core/top-bar";
import { Button } from "@/components/shadcn/button";
import { buildGetUserLeaguesQueryOptions } from "@/features/leagues/queries";
import type { User } from "@/features/users/types";
import { makeServerRequest } from "@/lib/query/server-requestor";
import { HydrationBoundary, QueryClient, dehydrate } from "@tanstack/react-query";
import { withAuth } from "@workos-inc/authkit-nextjs";
import Link from "next/link";
import { redirect } from "next/navigation";

function LandingPage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-8">
      <div className="max-w-2xl text-center">
        <h1 className="mb-4 text-5xl font-bold">GreenRats</h1>
        <p className="mb-8 text-xl text-muted-foreground">
          Pick one golfer per tournament. Compete with friends. Track your earnings throughout the
          PGA Tour season.
        </p>
        <div className="flex justify-center gap-4">
          <Link href="/login">
            <Button size="lg">Get Started</Button>
          </Link>
          <Link href="/rules">
            <Button variant="outline" size="lg">
              How to Play
            </Button>
          </Link>
        </div>
      </div>
    </main>
  );
}

interface LeaguePickerProps {
  displayName: string;
  dehydratedState: ReturnType<typeof dehydrate>;
}

function LeaguePicker({ displayName, dehydratedState }: LeaguePickerProps) {
  return (
    <div className="flex min-h-screen flex-col">
      <TopBar />
      <main className="flex-1 p-8">
        <HydrationBoundary state={dehydratedState}>
          <LeaguePickerContent displayName={displayName} />
        </HydrationBoundary>
      </main>
    </div>
  );
}

export default async function Home() {
  const { user: authUser } = await withAuth();

  if (!authUser) {
    return <LandingPage />;
  }

  let user: User;
  try {
    user = await makeServerRequest.get<User>("/api/v1/users/me");
  } catch {
    return <LandingPage />;
  }

  if (!user.display_name) {
    redirect("/onboarding");
  }

  const queryClient = new QueryClient();
  await queryClient.prefetchQuery(buildGetUserLeaguesQueryOptions(makeServerRequest));

  return (
    <LeaguePicker
      displayName={user.display_name || authUser.email || "User"}
      dehydratedState={dehydrate(queryClient)}
    />
  );
}
