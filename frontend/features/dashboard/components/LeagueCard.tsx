"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/shadcn/card";
import { LeagueMonogram } from "@/features/leagues/components";
import type { League } from "@/features/leagues/types";
import { CrownIcon, UsersIcon } from "lucide-react";
import Link from "next/link";

interface LeagueCardProps {
  league: League;
}

export function LeagueCard({ league }: LeagueCardProps) {
  return (
    <Link href={`/leagues/${league.id}`}>
      <Card className="hover:bg-muted/50 transition-colors">
        <CardHeader className="flex flex-row items-center gap-3 pb-2">
          <LeagueMonogram league={league} size={40} />
          <div className="flex-1">
            <CardTitle className="text-base font-medium">{league.name}</CardTitle>
            <div className="text-muted-foreground flex items-center gap-3 text-sm">
              <span className="flex items-center gap-1">
                <UsersIcon className="size-3" />
                {league.member_count ?? 0}
              </span>
              {league.role === "owner" && (
                <span className="flex items-center gap-1 text-amber-600">
                  <CrownIcon className="size-3" />
                  Commissioner
                </span>
              )}
            </div>
          </div>
        </CardHeader>
      </Card>
    </Link>
  );
}
