"use client";

import { useCommissionerActions } from "@/features/leagues/queries";
import type { CommissionerAction } from "@/features/leagues/types";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/shadcn/card";
import { Skeleton } from "@/components/shadcn/skeleton";
import { ClockIcon, KeyIcon, RefreshCwIcon, UsersIcon, ActivityIcon } from "lucide-react";

interface LeagueActivityProps {
  leagueId: string;
}

function getActionIcon(actionType: CommissionerAction["action_type"]) {
  switch (actionType) {
    case "pick_change":
      return <RefreshCwIcon className="size-4" />;
    case "join_code_reset":
      return <KeyIcon className="size-4" />;
    case "joining_disabled":
    case "joining_enabled":
      return <UsersIcon className="size-4" />;
    default:
      return <ClockIcon className="size-4" />;
  }
}

function formatDate(dateString: string) {
  const date = new Date(dateString);
  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

function getActionDescription(action: CommissionerAction): string {
  if (action.action_type === "pick_change" && action.metadata) {
    const oldGolfer = action.metadata.old_golfer_name as string | undefined;
    const newGolfer = action.metadata.new_golfer_name as string | undefined;
    const tournamentName = action.metadata.tournament_name as string | undefined;
    if (oldGolfer && newGolfer && action.affected_user_name) {
      return `Changed ${action.affected_user_name}'s pick from ${oldGolfer} to ${newGolfer}${tournamentName ? ` for ${tournamentName}` : ""}`;
    }
  }
  return action.description;
}

function ActionItem({ action }: { action: CommissionerAction }) {
  return (
    <div className="flex items-start gap-3 py-3">
      <div className="mt-0.5 rounded-full bg-muted p-2">{getActionIcon(action.action_type)}</div>
      <div className="min-w-0 flex-1">
        <p className="text-sm">{getActionDescription(action)}</p>
        <p className="text-xs text-muted-foreground">
          {formatDate(action.created_at)}
          {action.commissioner_name && ` by ${action.commissioner_name}`}
        </p>
      </div>
    </div>
  );
}

function EmptyState() {
  return (
    <Card>
      <CardContent className="flex flex-col items-center justify-center py-12 text-center">
        <div className="mb-4 rounded-full bg-muted p-4">
          <ActivityIcon className="size-8 text-muted-foreground" />
        </div>
        <h3 className="mb-2 text-lg font-medium">No activity yet</h3>
        <p className="max-w-sm text-sm text-muted-foreground">
          This page tracks commissioner changes to picks and league settings. Activity will appear
          here when commissioners make changes.
        </p>
      </CardContent>
    </Card>
  );
}

export function LeagueActivity({ leagueId }: LeagueActivityProps) {
  const { data, isLoading, error } = useCommissionerActions(leagueId);

  if (isLoading) {
    return (
      <Card>
        <CardHeader className="pb-3">
          <Skeleton className="h-5 w-32" />
          <Skeleton className="h-4 w-48" />
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="flex items-start gap-3">
                <Skeleton className="size-8 rounded-full" />
                <div className="flex-1 space-y-1">
                  <Skeleton className="h-4 w-3/4" />
                  <Skeleton className="h-3 w-1/2" />
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground">Failed to load activity</p>
        </CardContent>
      </Card>
    );
  }

  if (!data?.actions.length) {
    return <EmptyState />;
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base">League Activity</CardTitle>
        <CardDescription>Commissioner actions and changes</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="divide-y">
          {data.actions.map((action) => (
            <ActionItem key={action.id} action={action} />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
