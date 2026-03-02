"use client";

import { DashboardCard } from "./DashboardCard";
import { Button } from "@/components/shadcn/button";
import {
  Item,
  ItemContent,
  ItemFooter,
  ItemGroup,
  ItemHeader,
  ItemMedia,
  ItemSeparator,
  ItemTitle,
} from "@/components/shadcn/item";
import { useCommissionerActions } from "@/features/leagues/queries";
import type { CommissionerAction } from "@/features/leagues/types";
import { ActivityIcon, ClockIcon, KeyIcon, RefreshCwIcon, UsersIcon } from "lucide-react";
import Link from "next/link";
import { Fragment } from "react";

interface AuditCardProps {
  leagueId: string;
}

function getActionIcon(actionType: CommissionerAction["action_type"]) {
  switch (actionType) {
    case "pick_change":
      return <RefreshCwIcon className="size-3" />;
    case "join_code_reset":
      return <KeyIcon className="size-3" />;
    case "joining_disabled":
    case "joining_enabled":
      return <UsersIcon className="size-3" />;
    default:
      return <ClockIcon className="size-3" />;
  }
}

function formatDate(dateString: string) {
  const date = new Date(dateString);
  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
  });
}

function getActionDescription(action: CommissionerAction): string {
  if (action.action_type === "pick_change" && action.metadata) {
    const oldGolfer = action.metadata.old_golfer_name as string | undefined;
    const newGolfer = action.metadata.new_golfer_name as string | undefined;
    const tournamentName = action.metadata.tournament_name as string | undefined;
    if (oldGolfer && newGolfer) {
      return `Changed pick from ${oldGolfer} to ${newGolfer}${tournamentName ? ` for ${tournamentName}` : ""}`;
    }
  }
  return action.description;
}

export function AuditCard({ leagueId }: AuditCardProps) {
  const { data, isLoading, error } = useCommissionerActions(leagueId);

  const action = (
    <Button asChild variant="ghost" size="sm">
      <Link href={`/${leagueId}/audit`}>View all</Link>
    </Button>
  );

  if (error) {
    return (
      <DashboardCard
        title="Commissioner Activity"
        icon={<ActivityIcon className="size-4" />}
        action={action}
      >
        <div className="text-sm text-destructive">Failed to load activity</div>
      </DashboardCard>
    );
  }

  if (!data?.actions.length && !isLoading) {
    return (
      <DashboardCard
        title="Commissioner Activity"
        icon={<ActivityIcon className="size-4" />}
        action={action}
      >
        <p className="text-sm text-muted-foreground">No recent activity</p>
      </DashboardCard>
    );
  }

  const recentActions = data?.actions.slice(0, 5) ?? [];

  return (
    <DashboardCard
      title="Commissioner Activity"
      icon={<ActivityIcon className="size-4" />}
      action={action}
      isLoading={isLoading}
    >
      <ItemGroup className="gap-2">
        {recentActions.map((action, index) => (
          <Fragment key={action.id}>
            <Item size="sm" className="py-2">
              <ItemMedia variant="icon">{getActionIcon(action.action_type)}</ItemMedia>
              <ItemContent className="min-w-0">
                <ItemHeader>
                  <ItemTitle className="line-clamp-3 w-auto">{getActionDescription(action)}</ItemTitle>
                  <span className="text-xs text-muted-foreground">
                    {formatDate(action.created_at)}
                  </span>
                </ItemHeader>
                <ItemFooter className="text-xs text-muted-foreground">
                  <span>
                    Affected: {action.affected_user_name ? action.affected_user_name : "--"}
                  </span>
                  <span>By: {action.commissioner_name ? action.commissioner_name : "--"}</span>
                </ItemFooter>
              </ItemContent>
            </Item>
            {index < recentActions.length - 1 && <ItemSeparator />}
          </Fragment>
        ))}
      </ItemGroup>
    </DashboardCard>
  );
}
