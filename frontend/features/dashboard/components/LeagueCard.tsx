"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/shadcn/card";
import { LeagueMonogram } from "@/features/leagues/components";
import type { League } from "@/features/leagues/types";
import { formatDistanceToNow } from "date-fns";
import { CalendarIcon, CrownIcon, UserIcon, UsersIcon } from "lucide-react";
import Link from "next/link";

interface LeagueCardProps {
  league: League;
}

export function LeagueCard({ league }: LeagueCardProps) {
  return (
    <Link href={`/${league.id}`}>
      <Card className="transition-colors hover:bg-muted/50">
        <CardHeader className="flex flex-row items-center gap-3 pb-2">
          <LeagueMonogram league={league} size={40} />
          <div className="min-w-0 flex-1">
            <CardTitle className="truncate text-base font-medium">{league.name}</CardTitle>
            <div className="flex items-center gap-3 text-sm text-muted-foreground">
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
        {(league.recent_pick || league.next_deadline) && (
          <CardContent className="border-t pt-3">
            <div className="space-y-1 text-xs text-muted-foreground">
              {league.recent_pick && (
                <div className="flex items-center gap-2">
                  <UserIcon className="size-3" />
                  <span className="truncate">
                    {league.recent_pick.golfer_name} ({league.recent_pick.tournament_name})
                  </span>
                </div>
              )}
              {league.next_deadline && (
                <div className="flex items-center gap-2">
                  <CalendarIcon className="size-3" />
                  <span>
                    {league.next_deadline.tournament_name} -{" "}
                    {formatDistanceToNow(new Date(league.next_deadline.deadline), {
                      addSuffix: true,
                    })}
                  </span>
                </div>
              )}
            </div>
          </CardContent>
        )}
      </Card>
    </Link>
  );
}
