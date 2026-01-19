"use client";

import { useState } from "react";

import { useActiveTournament, useTournaments } from "@/features/tournaments/queries";
import { cn } from "@/lib/utils";
import { format } from "date-fns";
import { ChevronDownIcon, ChevronUpIcon } from "lucide-react";

function formatDateRange(startDate: Date, endDate: Date): string {
  const startMonth = format(startDate, "MMM");
  const endMonth = format(endDate, "MMM");
  const startDay = format(startDate, "d");
  const endDay = format(endDate, "d");

  if (startMonth === endMonth) {
    return `${startMonth} ${startDay}-${endDay}`;
  }
  return `${startMonth} ${startDay}-${endMonth} ${endDay}`;
}

export function TournamentCalendarRow() {
  const [isExpanded, setIsExpanded] = useState(false);
  const { data: activeTournamentData } = useActiveTournament();
  const { data: upcomingData } = useTournaments({ status: "upcoming", limit: 6 });

  const activeTournament = activeTournamentData?.tournament;
  const upcomingTournaments = upcomingData?.tournaments ?? [];

  const filteredUpcoming = activeTournament
    ? upcomingTournaments.filter((t) => t.id !== activeTournament.id)
    : upcomingTournaments;

  const primaryTournament = activeTournament ?? filteredUpcoming[0];
  const remainingTournaments = activeTournament ? filteredUpcoming : filteredUpcoming.slice(1);

  if (!primaryTournament) {
    return null;
  }

  const isActive = primaryTournament.status === "active";
  const startDate = new Date(primaryTournament.start_date);
  const endDate = new Date(primaryTournament.end_date);

  return (
    <div className="border-border bg-card w-full rounded-lg border">
      <button
        type="button"
        onClick={() => setIsExpanded(!isExpanded)}
        className={cn(
          "flex w-full items-center justify-between px-4 py-3",
          remainingTournaments.length === 0 && "cursor-default",
        )}
        disabled={remainingTournaments.length === 0}
      >
        <div className="flex items-center gap-3">
          {isActive && <span className="bg-primary size-2 animate-pulse rounded-full" />}
          <span className="font-medium">{primaryTournament.name}</span>
          <span className="text-muted-foreground text-sm">{formatDateRange(startDate, endDate)}</span>
        </div>
        {remainingTournaments.length > 0 && (
          <div className="text-muted-foreground flex items-center gap-1 text-sm">
            <span>{remainingTournaments.length} more</span>
            {isExpanded ? <ChevronUpIcon className="size-4" /> : <ChevronDownIcon className="size-4" />}
          </div>
        )}
      </button>

      {isExpanded && remainingTournaments.length > 0 && (
        <div className="border-border border-t px-4 py-2">
          {remainingTournaments.map((tournament) => {
            const tStart = new Date(tournament.start_date);
            const tEnd = new Date(tournament.end_date);
            return (
              <div key={tournament.id} className="text-muted-foreground flex items-center justify-between py-2 text-sm">
                <span>{tournament.name}</span>
                <span>{formatDateRange(tStart, tEnd)}</span>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
