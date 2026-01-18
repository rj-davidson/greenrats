import { redirect } from "next/navigation";

import { Button } from "@/components/shadcn/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/shadcn/card";
import { LeaguesSection } from "@/features/leagues/components";
import type { User } from "@/features/users/types";
import { makeServerRequest } from "@/lib/query/server-requestor";

export default async function DashboardPage() {
  // Fetch DB user (auto-creates if new WorkOS user)
  let user: User | null = null;
  try {
    user = await makeServerRequest.get<User>("/api/v1/users/me");
  } catch (error) {
    console.error("Failed to fetch user:", error);
  }

  // Redirect to onboarding if user doesn't have a display name
  if (user && !user.display_name) {
    redirect("/onboarding");
  }

  const displayName = user?.display_name || user?.email || "User";

  return (
    <main className="container mx-auto p-8">
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
    </main>
  );
}
