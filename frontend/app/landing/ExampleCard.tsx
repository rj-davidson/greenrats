import { Badge } from "@/components/shadcn/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/shadcn/card";
import { cn } from "@/lib/utils";
import type { ReactNode } from "react";

interface ExampleCardProps {
  title: string;
  icon?: ReactNode;
  className?: string;
  children?: ReactNode;
}

export function ExampleCard({ title, icon, className, children }: ExampleCardProps) {
  return (
    <Card className={cn("flex flex-col gap-3 py-4 opacity-90", className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 px-4 pb-0">
        <CardTitle className="flex items-center gap-2 text-sm font-medium">
          {icon}
          {title}
        </CardTitle>
        <Badge variant="outline" className="text-xs">
          Example
        </Badge>
      </CardHeader>
      <CardContent className="flex-1 px-4">{children}</CardContent>
    </Card>
  );
}
