"use client";

import { Button } from "@/components/shadcn/button";
import { Card, CardContent } from "@/components/shadcn/card";
import { Skeleton } from "@/components/shadcn/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/shadcn/table";
import { useCommissionerActions } from "@/features/leagues/queries";
import type { CommissionerAction } from "@/features/leagues/types";
import {
  ActivityIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  ClockIcon,
  KeyIcon,
  RefreshCwIcon,
  UsersIcon,
} from "lucide-react";
import { useState } from "react";

interface LeagueActivityProps {
  leagueId: string;
}

const PAGE_SIZE = 20;

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
    if (oldGolfer && newGolfer) {
      return `Changed pick from ${oldGolfer} to ${newGolfer}${tournamentName ? ` for ${tournamentName}` : ""}`;
    }
  }
  return action.description;
}

function ActivitySkeleton() {
  return (
    <div className="space-y-2">
      {Array.from({ length: 8 }).map((_, i) => (
        <Skeleton key={i} className="h-12 w-full" />
      ))}
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
          Activity will appear here when commissioners make changes to picks or league settings.
        </p>
      </CardContent>
    </Card>
  );
}

export function LeagueActivity({ leagueId }: LeagueActivityProps) {
  const [page, setPage] = useState(0);
  const { data, isLoading, error } = useCommissionerActions(leagueId);

  if (isLoading) {
    return <ActivitySkeleton />;
  }

  if (error) {
    return <div className="text-destructive">Failed to load activity</div>;
  }

  if (!data?.actions.length) {
    return <EmptyState />;
  }

  const totalPages = Math.ceil(data.actions.length / PAGE_SIZE);
  const showPagination = data.actions.length > PAGE_SIZE;
  const paginatedActions = data.actions.slice(page * PAGE_SIZE, (page + 1) * PAGE_SIZE);

  return (
    <div className="space-y-4">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-12">Type</TableHead>
            <TableHead>Action</TableHead>
            <TableHead>Affected</TableHead>
            <TableHead>By</TableHead>
            <TableHead className="text-right">Date</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {paginatedActions.map((action) => (
            <TableRow key={action.id}>
              <TableCell>
                <div className="w-fit rounded-full bg-muted p-2">
                  {getActionIcon(action.action_type)}
                </div>
              </TableCell>
              <TableCell>{getActionDescription(action)}</TableCell>
              <TableCell className="text-muted-foreground">
                {action.affected_user_name || "--"}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {action.commissioner_name || "--"}
              </TableCell>
              <TableCell className="text-right text-muted-foreground">
                {formatDate(action.created_at)}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>

      {showPagination && (
        <div className="flex items-center justify-between">
          <span className="text-sm text-muted-foreground">
            Page {page + 1} of {totalPages}
          </span>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => p - 1)}
              disabled={page === 0}
            >
              <ChevronLeftIcon className="size-4" />
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => p + 1)}
              disabled={page >= totalPages - 1}
            >
              Next
              <ChevronRightIcon className="size-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
