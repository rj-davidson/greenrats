"use client";

import { Card, CardContent, CardHeader } from "@/components/shadcn/card";
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemMedia,
  ItemTitle,
} from "@/components/shadcn/item";
import { Skeleton } from "@/components/shadcn/skeleton";
import { cn } from "@/lib/utils";
import type { ReactNode } from "react";

interface DashboardCardProps {
  title: string;
  description?: string;
  icon?: ReactNode;
  action?: ReactNode;
  isLoading?: boolean;
  className?: string;
  children?: ReactNode;
}

export function DashboardCard({
  title,
  description,
  icon,
  action,
  isLoading,
  className,
  children,
}: DashboardCardProps) {
  return (
    <Card className={cn("flex flex-col gap-3 py-4", className)}>
      <CardHeader className="px-4 pb-0">
        <Item size="sm" className="p-0">
          {icon && <ItemMedia>{icon}</ItemMedia>}
          <ItemContent className="gap-0.5">
            <ItemTitle>{title}</ItemTitle>
            {description && <ItemDescription>{description}</ItemDescription>}
          </ItemContent>
          {action && <ItemActions>{action}</ItemActions>}
        </Item>
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
