"use client";

import type { AvailableGolfer } from "../types";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/shadcn/avatar";
import { Badge } from "@/components/shadcn/badge";
import { cn } from "@/lib/utils";

interface GolferCardProps {
  golfer: AvailableGolfer;
  selected?: boolean;
  onClick?: () => void;
}

export function GolferCard({ golfer, selected, onClick }: GolferCardProps) {
  const initials = golfer.name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();

  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "flex w-full items-center gap-3 rounded-lg border p-3 text-left transition-colors",
        "hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
        selected && "border-primary bg-primary/5",
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
            <span className="text-muted-foreground shrink-0 text-xs">{golfer.country_code}</span>
          )}
        </div>
        <div className="text-muted-foreground flex items-center gap-2 text-sm">
          {golfer.owgr && golfer.owgr > 0 && <span>OWGR #{golfer.owgr}</span>}
          {golfer.country && <span className="truncate">{golfer.country}</span>}
        </div>
      </div>
      {selected && (
        <Badge variant="default" className="shrink-0">
          Selected
        </Badge>
      )}
    </button>
  );
}
