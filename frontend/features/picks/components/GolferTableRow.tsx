"use client";

import type { AvailableGolfer } from "@/features/picks/types";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/shadcn/avatar";
import { Button } from "@/components/shadcn/button";
import { TableCell, TableRow } from "@/components/shadcn/table";
import { cn } from "@/lib/utils";

interface GolferTableRowProps {
  golfer: AvailableGolfer;
  onSelect: (golfer: AvailableGolfer) => void;
}

export function GolferTableRow({ golfer, onSelect }: GolferTableRowProps) {
  const initials = golfer.name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();

  const isUsed = golfer.is_used ?? false;

  return (
    <TableRow className={cn(isUsed && "opacity-50")}>
      <TableCell>
        <div className="flex items-center gap-3">
          <Avatar className="size-8">
            {golfer.image_url && <AvatarImage src={golfer.image_url} alt={golfer.name} />}
            <AvatarFallback className="text-xs">{initials}</AvatarFallback>
          </Avatar>
          <div className="min-w-0">
            <div className="truncate font-medium">{golfer.name}</div>
            <div className="text-xs text-muted-foreground md:hidden">{golfer.country_code}</div>
          </div>
        </div>
      </TableCell>
      <TableCell className="hidden md:table-cell">{golfer.country_code}</TableCell>
      <TableCell className="text-right tabular-nums">
        {golfer.owgr && golfer.owgr > 0 ? `#${golfer.owgr}` : "—"}
      </TableCell>
      <TableCell className="text-right">
        {isUsed ? (
          <span className="text-sm text-muted-foreground">{golfer.used_for_tournament}</span>
        ) : (
          <Button size="sm" onClick={() => onSelect(golfer)}>
            Select
          </Button>
        )}
      </TableCell>
    </TableRow>
  );
}
