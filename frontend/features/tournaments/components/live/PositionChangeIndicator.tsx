"use client";

import { cn } from "@/lib/utils";
import { MoveDownIcon, MoveUpIcon } from "lucide-react";

type PositionChangeIndicatorProps = {
  change: number | null | undefined;
  className?: string;
};

export function PositionChangeIndicator({ change, className }: PositionChangeIndicatorProps) {
  if (change == null || change === 0) {
    return <span className={cn("text-muted-foreground", className)}>-</span>;
  }

  if (change > 0) {
    return (
      <span
        className={cn(
          "flex flex-nowrap items-center gap-0 text-green-600 dark:text-green-400",
          className,
        )}
      >
        <MoveUpIcon className="inline size-3" />
        {change}
      </span>
    );
  }

  return (
    <span
      className={cn(
        "flex flex-nowrap items-center gap-0 text-red-600 dark:text-red-400",
        className,
      )}
    >
      <MoveDownIcon className="inline size-3" />
      {Math.abs(change)}
    </span>
  );
}
