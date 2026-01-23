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
import { useUserPicks } from "@/features/picks/queries";
import { HistoryIcon } from "lucide-react";
import Link from "next/link";
import { useMemo } from "react";

interface PickHistoryCardProps {
  leagueId: string;
}

function formatEarnings(amount: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(amount);
}

export function PickHistoryCard({ leagueId }: PickHistoryCardProps) {
  const { data, isLoading } = useUserPicks(leagueId);

  const { picks, totalEarnings } = useMemo(() => {
    if (!data?.picks) return { picks: [], totalEarnings: 0 };

    const sortedPicks = [...data.picks].sort((a, b) => {
      const aTime = Date.parse(a.created_at);
      const bTime = Date.parse(b.created_at);
      if (Number.isNaN(aTime) || Number.isNaN(bTime)) {
        return b.created_at.localeCompare(a.created_at);
      }
      return bTime - aTime;
    });

    const total = sortedPicks.reduce((sum, p) => sum + (p.golfer_earnings ?? 0), 0);

    return { picks: sortedPicks, totalEarnings: total };
  }, [data]);

  const action = (
    <Button asChild variant="ghost" size="sm">
      <Link href={`/${leagueId}/standings`}>View all</Link>
    </Button>
  );

  if (!picks.length && !isLoading) {
    return (
      <DashboardCard title="Pick History" icon={<HistoryIcon className="size-4" />} action={action}>
        <p className="text-sm text-muted-foreground">You haven&apos;t made any picks yet.</p>
      </DashboardCard>
    );
  }

  return (
    <DashboardCard
      title="Pick History"
      icon={<HistoryIcon className="size-4" />}
      action={action}
      isLoading={isLoading}
    >
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Tournament</TableHead>
            <TableHead>Golfer</TableHead>
            <TableHead className="text-right">Earnings</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {picks.map((pick) => (
            <TableRow key={pick.id}>
              <TableCell className="max-w-32 truncate text-sm">{pick.tournament_name}</TableCell>
              <TableCell className="text-sm">{pick.golfer_name}</TableCell>
              <TableCell className="text-right text-sm">
                {pick.golfer_earnings !== undefined ? formatEarnings(pick.golfer_earnings) : "-"}
              </TableCell>
            </TableRow>
          ))}
          {picks.length > 0 && (
            <TableRow className="border-t-2">
              <TableCell colSpan={2} className="font-semibold">
                Total
              </TableCell>
              <TableCell className="text-right font-semibold">
                {formatEarnings(totalEarnings)}
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </DashboardCard>
  );
}
