"use client";

import { useUserLeagues } from "@/features/leagues/queries";
import type { League } from "@/features/leagues/types";
import { Badge } from "@/components/shadcn/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/shadcn/card";

function RoleBadge({ role }: { role: string }) {
  const isOwner = role === "Owner";
  return <Badge variant={isOwner ? "default" : "secondary"}>{role}</Badge>;
}

function LeagueCard({ league }: { league: League }) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between">
          <CardTitle className="text-lg">{league.name}</CardTitle>
          {league.role && <RoleBadge role={league.role} />}
        </div>
        <CardDescription>Season {league.season_year}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-2 text-sm">
          <span className="text-muted-foreground">Join Code:</span>
          <code className="rounded bg-muted px-2 py-1 font-mono">{league.code}</code>
        </div>
      </CardContent>
    </Card>
  );
}

export function LeaguesList() {
  const { data, isLoading, error } = useUserLeagues();

  if (isLoading) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {[1, 2, 3].map((i) => (
          <Card key={i} className="animate-pulse">
            <CardHeader>
              <div className="h-5 w-32 rounded bg-muted" />
              <div className="h-4 w-20 rounded bg-muted" />
            </CardHeader>
            <CardContent>
              <div className="h-4 w-24 rounded bg-muted" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground">Failed to load leagues. Please try again.</p>
        </CardContent>
      </Card>
    );
  }

  if (!data?.leagues.length) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground">
            You haven&apos;t joined any leagues yet. Create one to get started!
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {data.leagues.map((league) => (
        <LeagueCard key={league.id} league={league} />
      ))}
    </div>
  );
}
