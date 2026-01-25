"use client";

import { cn } from "@/lib/utils";
import { ChevronDown, ChevronUp } from "lucide-react";

interface PositionChangeIndicatorProps {
  change: number | null | undefined;
  className?: string;
}

export function PositionChangeIndicator({ change, className }: PositionChangeIndicatorProps) {
  if (change == null || change === 0) {
    return <span className={cn("text-muted-foreground", className)}>-</span>;
  }

  if (change > 0) {
    return (
      <span className={cn("text-green-600 dark:text-green-400", className)}>
        <ChevronUp className="inline size-3" />
        {change}
      </span>
    );
  }

  return (
    <span className={cn("text-red-600 dark:text-red-400", className)}>
      <ChevronDown className="inline size-3" />
      {Math.abs(change)}
    </span>
  );
}
