"use client";

import { Button } from "@/components/shadcn/button";
import { TableCell, TableRow } from "@/components/shadcn/table";
import { GolferDetailPanel } from "@/features/picks/components/GolferDetailPanel";
import type { PickFieldEntry, PickWindowState } from "@/features/picks/types";
import { cn } from "@/lib/utils";
import { CheckIcon, ChevronUpIcon } from "lucide-react";
import { Fragment } from "react";

interface PickFieldRowProps {
  entry: PickFieldEntry;
  isExpanded: boolean;
  isCurrentPick: boolean;
  pickWindowState: PickWindowState;
  onToggle: () => void;
  onSelect: () => void;
}

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(amount);
}

export function PickFieldRow({
  entry,
  isExpanded,
  isCurrentPick,
  pickWindowState,
  onToggle,
  onSelect,
}: PickFieldRowProps) {
  const isUsed = entry.is_used && !isCurrentPick;
  const canPick = pickWindowState === "open" && !isUsed;

  return (
    <Fragment>
      <TableRow
        className={cn(
          "cursor-pointer",
          isUsed && "opacity-50",
          isCurrentPick && "bg-primary/5"
        )}
        onClick={onToggle}
      >
        <TableCell className="w-0" onClick={(e) => e.stopPropagation()}>
          {isUsed ? (
            <Button size="sm" variant="ghost" disabled className="w-14">
              Used
            </Button>
          ) : isCurrentPick ? (
            <Button size="sm" onClick={onSelect} disabled={!canPick} className="w-14">
              <CheckIcon className="size-4" />
            </Button>
          ) : canPick ? (
            <Button size="sm" variant="outline" onClick={onSelect} className="w-14">
              Pick
            </Button>
          ) : (
            <Button size="sm" variant="ghost" disabled className="w-14">
              -
            </Button>
          )}
        </TableCell>
        <TableCell>{entry.golfer_name}</TableCell>
        <TableCell className="text-right tabular-nums">
          {entry.owgr && entry.owgr > 0 ? `#${entry.owgr}` : "-"}
        </TableCell>
        <TableCell className="hidden text-right tabular-nums sm:table-cell">
          {entry.season_earnings ? formatCurrency(entry.season_earnings) : "-"}
        </TableCell>
      </TableRow>
      {isExpanded && (
        <TableRow className="hover:bg-transparent">
          <TableCell colSpan={4} className="p-0">
            <div className="border-t bg-muted/30">
              <GolferDetailPanel stats={entry.season_stats} bio={entry.bio} />
              <div className="flex justify-center border-t py-2">
                <Button variant="ghost" size="sm" onClick={onToggle} className="gap-1">
                  <ChevronUpIcon className="size-4" />
                </Button>
              </div>
            </div>
          </TableCell>
        </TableRow>
      )}
    </Fragment>
  );
}
