"use client";

import type { AvailableGolfer } from "@/features/picks/types";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/shadcn/avatar";
import { Badge } from "@/components/shadcn/badge";
import { cn } from "@/lib/utils";

interface GolferCardProps {
  golfer: AvailableGolfer;
  selected?: boolean;
  onClick?: () => void;
  disabled?: boolean;
  disabledReason?: string;
}

export function GolferCard({
  golfer,
  selected,
  onClick,
  disabled,
  disabledReason,
}: GolferCardProps) {
  const initials = golfer.name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();

  return (
    <button
      type="button"
      onClick={disabled ? undefined : onClick}
      disabled={disabled}
      className={cn(
        "flex w-full items-center gap-3 rounded-lg border p-3 text-left transition-colors",
        !disabled && "hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none",
        selected && "border-primary bg-primary/5",
        disabled && "cursor-not-allowed opacity-50",
      )}
    >
      <Avatar className="size-10">
        {golfer.image_url && <AvatarImage src={golfer.image_url} alt={golfer.name} />}
        <AvatarFallback className="text-xs">{initials}</AvatarFallback>
      </Avatar>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="truncate font-medium">{golfer.name}</span>
          {golfer.country_code && (
            <span className="shrink-0 text-xs text-muted-foreground">{golfer.country_code}</span>
          )}
        </div>
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          {golfer.owgr && golfer.owgr > 0 && <span>OWGR #{golfer.owgr}</span>}
          {golfer.country && <span className="truncate">{golfer.country}</span>}
        </div>
        {disabled && disabledReason && (
          <div className="mt-1 text-xs text-muted-foreground">{disabledReason}</div>
        )}
      </div>
      {selected && (
        <Badge variant="default" className="shrink-0">
          Selected
        </Badge>
      )}
    </button>
  );
}
