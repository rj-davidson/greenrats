"use client";

import { Badge } from "@/components/shadcn/badge";
import { Item, ItemContent, ItemMedia } from "@/components/shadcn/item";
import { Skeleton } from "@/components/shadcn/skeleton";
import {
  formatPickWindowDate,
  getPickWindowCountdown,
  getPickWindowState,
} from "@/features/picks/utils";
import { useCurrentTournament } from "@/features/tournaments/queries";
import type { Tournament } from "@/features/tournaments/types";
import { format } from "date-fns";
import { CalendarIcon, ClockIcon, MapPinIcon, TrophyIcon } from "lucide-react";

function formatDateRange(startDate: Date, endDate: Date): string {
  const startMonth = format(startDate, "MMM");
  const endMonth = format(endDate, "MMM");
  const startDay = format(startDate, "d");
  const endDay = format(endDate, "d");
  const year = format(endDate, "yyyy");

  if (startMonth === endMonth) {
    return `${startMonth} ${startDay}-${endDay}, ${year}`;
  }
  return `${startMonth} ${startDay} - ${endMonth} ${endDay}, ${year}`;
}

function formatPurse(purse: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(purse);
}

function StatusBadge({ status }: { status: Tournament["status"] }) {
  switch (status) {
    case "active":
      return (
        <Badge variant="outline" className="animate-shimmer border-primary text-primary">
          LIVE
        </Badge>
      );
    case "completed":
      return <Badge variant="secondary">Final</Badge>;
    case "upcoming":
      return <Badge variant="outline">Upcoming</Badge>;
  }
}

function DetailItem({
  icon: Icon,
  label,
  children,
}: {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  children: React.ReactNode;
}) {
  return (
    <Item size="sm" className="py-1.5">
      <ItemMedia>
        <Icon className="size-4 text-muted-foreground" />
      </ItemMedia>
      <ItemContent>
        <span className="text-sm">
          <span className="text-muted-foreground">{label}:</span>{" "}
          {children}
        </span>
      </ItemContent>
    </Item>
  );
}

function PickWindowInfo({ tournament }: { tournament: Tournament }) {
  const pickWindowState = getPickWindowState(tournament);
  const { countdown } = getPickWindowCountdown(tournament);

  if (pickWindowState === "closed") {
    return null;
  }

  if (pickWindowState === "open" && tournament.pick_window_closes_at) {
    return (
      <DetailItem icon={ClockIcon} label="Picks close">
        <span className="font-medium text-primary">{countdown}</span>
        <span className="text-muted-foreground">
          {" "}({formatPickWindowDate(tournament.pick_window_closes_at)})
        </span>
      </DetailItem>
    );
  }

  if (pickWindowState === "not_open" && tournament.pick_window_opens_at) {
    return (
      <DetailItem icon={ClockIcon} label="Picks open">
        {countdown} ({formatPickWindowDate(tournament.pick_window_opens_at)})
      </DetailItem>
    );
  }

  return null;
}

function TournamentDetails({ tournament }: { tournament: Tournament }) {
  const startDate = new Date(tournament.start_date);
  const endDate = new Date(tournament.end_date);

  return (
    <div className="flex flex-col gap-0.5">
      <DetailItem icon={CalendarIcon} label="Dates">
        {formatDateRange(startDate, endDate)}
      </DetailItem>

      {tournament.course && (
        <DetailItem icon={MapPinIcon} label="Course">
          {tournament.course}
        </DetailItem>
      )}

      {tournament.city && tournament.state && (
        <DetailItem icon={MapPinIcon} label="Location">
          {tournament.city}, {tournament.state}
        </DetailItem>
      )}

      {tournament.purse && (
        <DetailItem icon={TrophyIcon} label="Purse">
          {formatPurse(tournament.purse)}
        </DetailItem>
      )}

      <PickWindowInfo tournament={tournament} />
    </div>
  );
}

export function PrimaryTournamentCard() {
  const { tournament, isLoading } = useCurrentTournament();

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-3">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-6 w-20" />
        </div>
        <div className="space-y-2">
          <Skeleton className="h-5 w-48" />
          <Skeleton className="h-5 w-56" />
          <Skeleton className="h-5 w-40" />
        </div>
      </div>
    );
  }

  if (!tournament) {
    return null;
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-3">
        <h1 className="text-3xl font-bold">{tournament.name}</h1>
        <StatusBadge status={tournament.status} />
      </div>
      <TournamentDetails tournament={tournament} />
    </div>
  );
}
