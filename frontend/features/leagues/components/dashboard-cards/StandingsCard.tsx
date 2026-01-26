"use client";

import { DashboardCard } from "./DashboardCard";
import { Button } from "@/components/shadcn/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/shadcn/table";
import { useLeagueLeaderboard } from "@/features/leaderboards/queries";
import { useCurrentUser } from "@/features/users/queries";
import { cn } from "@/lib/utils";
import { UsersIcon } from "lucide-react";
import Link from "next/link";

interface StandingsCardProps {
  leagueId: string;
}

function formatEarnings(amount: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(amount);
}

function getOrdinalSuffix(n: number): string {
  const s = ["th", "st", "nd", "rd"];
  const v = n % 100;
  return n + (s[(v - 20) % 10] || s[v] || s[0]);
}

export function StandingsCard({ leagueId }: StandingsCardProps) {
  const { data, isLoading, error } = useLeagueLeaderboard(leagueId);
  const { data: currentUser } = useCurrentUser();

  const action = (
    <Button asChild variant="ghost" size="sm">
      <Link href={`/${leagueId}/standings`}>View all</Link>
    </Button>
  );

  if (error) {
    return (
      <DashboardCard title="Standings" icon={<UsersIcon className="size-4" />} action={action}>
        <div className="text-destructive">Failed to load standings</div>
      </DashboardCard>
    );
  }

  if (!data?.entries.length && !isLoading) {
    return (
      <DashboardCard title="Standings" icon={<UsersIcon className="size-4" />} action={action}>
        <div className="py-4 text-center text-sm text-muted-foreground">
          No picks have been made yet.
        </div>
      </DashboardCard>
    );
  }

  const top5 = data?.entries.slice(0, 5) ?? [];
  const userEntry = data?.entries.find((e) => e.user_id === currentUser?.id);
  const userInTop5 = userEntry && userEntry.rank <= 5;
  const displayEntries = userInTop5 || !userEntry ? top5 : (data?.entries.slice(0, 4) ?? []);
  const showUserBubble = userEntry && !userInTop5;

  return (
    <DashboardCard
      title="Standings"
      icon={<UsersIcon className="size-4" />}
      action={action}
      isLoading={isLoading}
    >
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-6">#</TableHead>
            <TableHead>Player</TableHead>
            <TableHead className="text-right">Earnings</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {displayEntries.map((entry) => {
            const isCurrentUser = currentUser?.id === entry.user_id;
            return (
              <TableRow key={entry.user_id} className={cn(isCurrentUser && "bg-primary/5")}>
                <TableCell className="font-medium">{entry.rank}</TableCell>
                <TableCell>
                  {entry.display_name}
                  {isCurrentUser && (
                    <span className="ml-1 text-xs text-muted-foreground">(you)</span>
                  )}
                </TableCell>
                <TableCell className="text-right">{formatEarnings(entry.earnings)}</TableCell>
              </TableRow>
            );
          })}
          {showUserBubble && (
            <>
              <TableRow>
                <TableCell colSpan={3} className="py-1 text-center text-xs text-muted-foreground">
                  ...
                </TableCell>
              </TableRow>
              <TableRow className="bg-primary/5">
                <TableCell className="font-medium">{getOrdinalSuffix(userEntry.rank)}</TableCell>
                <TableCell>
                  {userEntry.display_name}
                  <span className="ml-1 text-xs text-muted-foreground">(you)</span>
                </TableCell>
                <TableCell className="text-right">{formatEarnings(userEntry.earnings)}</TableCell>
              </TableRow>
            </>
          )}
        </TableBody>
      </Table>
    </DashboardCard>
  );
}
