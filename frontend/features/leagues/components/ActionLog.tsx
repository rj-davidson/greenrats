"use client";

import { useCommissionerActions } from "../queries";
import type { CommissionerAction } from "../types";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/shadcn/card";
import { Skeleton } from "@/components/shadcn/skeleton";
import { ClockIcon, KeyIcon, RefreshCwIcon, UsersIcon } from "lucide-react";

interface ActionLogProps {
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

function ActionItem({ action }: { action: CommissionerAction }) {
  return (
    <div className="flex items-start gap-3 py-2">
      <div className="bg-muted mt-0.5 rounded-full p-2">{getActionIcon(action.action_type)}</div>
      <div className="min-w-0 flex-1">
        <p className="text-sm">{action.description}</p>
        <p className="text-muted-foreground text-xs">
          {formatDate(action.created_at)}
          {action.commissioner_name && ` by ${action.commissioner_name}`}
        </p>
      </div>
    </div>
  );
}

export function ActionLog({ leagueId }: ActionLogProps) {
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
    return null;
  }

  if (!data?.actions.length) {
    return null;
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base">Recent Activity</CardTitle>
        <CardDescription>Commissioner actions for this league</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="divide-y">
          {data.actions.slice(0, 10).map((action) => (
            <ActionItem key={action.id} action={action} />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
