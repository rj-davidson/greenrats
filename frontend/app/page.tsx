import { redirect } from "next/navigation";
import Link from "next/link";
import { withAuth } from "@workos-inc/authkit-nextjs";

import { AppSidebar } from "@/components/core/app-sidebar";
import { Button } from "@/components/shadcn/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/shadcn/card";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/shadcn/sidebar";
import { LeaguesSection } from "@/features/leagues/components";
import type { User } from "@/features/users/types";
import { makeServerRequest } from "@/lib/query/server-requestor";

function LandingPage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-8">
      <div className="max-w-2xl text-center">
        <h1 className="mb-4 text-5xl font-bold">GreenRats</h1>
        <p className="text-muted-foreground mb-8 text-xl">
          Pick one golfer per tournament. Compete with friends. Track your earnings throughout the
          PGA Tour season.
        </p>
        <div className="flex justify-center gap-4">
          <Link href="/login">
            <Button size="lg">Get Started</Button>
          </Link>
        </div>
      </div>
    </main>
  );
}

interface DashboardHomeProps {
  displayName: string;
}

function DashboardHome({ displayName }: DashboardHomeProps) {
  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
        </header>
        <main className="flex-1 p-4">
          <div className="container mx-auto p-8">
            <div className="mb-8">
              <h1 className="mb-2 text-3xl font-bold">Dashboard</h1>
              <p className="text-muted-foreground">Welcome back, {displayName}</p>
            </div>

            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle>Active Tournament</CardTitle>
                  <CardDescription>Make your pick for this week</CardDescription>
                </CardHeader>
                <CardContent>
                  <Button className="w-full">View Tournament</Button>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Your Picks</CardTitle>
                  <CardDescription>Season picks history</CardDescription>
                </CardHeader>
                <CardContent>
                  <Button variant="outline" className="w-full">
                    View Picks
                  </Button>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Leaderboard</CardTitle>
                  <CardDescription>League standings</CardDescription>
                </CardHeader>
                <CardContent>
                  <Button variant="outline" className="w-full">
                    View Standings
                  </Button>
                </CardContent>
              </Card>
            </div>

            <div className="mt-8">
              <LeaguesSection />
            </div>
          </div>
        </main>
      </SidebarInset>
    </SidebarProvider>
  );
}

export default async function Home() {
  const { user: authUser } = await withAuth();

  if (!authUser) {
    return <LandingPage />;
  }

  let user: User | null = null;
  try {
    user = await makeServerRequest.get<User>("/api/v1/users/me");
  } catch (error) {
    console.error("Failed to fetch user:", error);
  }

  if (user && !user.display_name) {
    redirect("/onboarding");
  }

  const displayName = user?.display_name || user?.email || "User";

  return <DashboardHome displayName={displayName} />;
}
