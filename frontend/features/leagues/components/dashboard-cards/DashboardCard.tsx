"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/shadcn/card";
import { Skeleton } from "@/components/shadcn/skeleton";
import { cn } from "@/lib/utils";
import type { ReactNode } from "react";

interface DashboardCardProps {
  title: string;
  icon?: ReactNode;
  action?: ReactNode;
  isLoading?: boolean;
  className?: string;
  children?: ReactNode;
}

export function DashboardCard({
  title,
  icon,
  action,
  isLoading,
  className,
  children,
}: DashboardCardProps) {
  return (
    <Card className={cn("flex flex-col gap-3 py-4", className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 px-4 pb-0">
        <CardTitle className="flex items-center gap-2 text-sm font-medium">
          {icon}
          {title}
        </CardTitle>
        {action}
      </CardHeader>
      <CardContent className="flex-1 px-4">
        {isLoading ? (
          <div className="space-y-2">
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-3/4" />
          </div>
        ) : (
          children
        )}
      </CardContent>
    </Card>
  );
}
