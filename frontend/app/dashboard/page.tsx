import { withAuth } from "@workos-inc/authkit-nextjs";
import { Button } from "@/components/shadcn/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/shadcn/card";

export default async function DashboardPage() {
  const { user } = await withAuth({ ensureSignedIn: true });

  return (
    <main className="container mx-auto p-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Dashboard</h1>
        <p className="text-muted-foreground">
          Welcome back, {user.firstName || user.email}
        </p>
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
    </main>
  );
}
