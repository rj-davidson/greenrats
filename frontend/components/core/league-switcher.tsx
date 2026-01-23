"use client";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/shadcn/dropdown-menu";
import { SidebarMenu, SidebarMenuButton, SidebarMenuItem } from "@/components/shadcn/sidebar";
import { Skeleton } from "@/components/shadcn/skeleton";
import { LeagueMonogram } from "@/features/leagues/components";
import { useUserLeagues } from "@/features/leagues/queries";
import type { League } from "@/features/leagues/types";
import { ChevronsUpDownIcon, LayoutGridIcon } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useMemo } from "react";

interface LeagueSwitcherProps {
  currentLeague: Pick<League, "id" | "name">;
}

export function LeagueSwitcher({ currentLeague }: LeagueSwitcherProps) {
  const { data, isLoading } = useUserLeagues();
  const router = useRouter();

  const sortedLeagues = useMemo(() => {
    if (!data?.leagues) return [];
    return [...data.leagues].sort((a, b) => a.name.localeCompare(b.name));
  }, [data]);

  const otherLeagues = sortedLeagues.filter((l) => l.id !== currentLeague.id);

  if (isLoading) {
    return (
      <SidebarMenu>
        <SidebarMenuItem>
          <div className="flex items-center gap-2 px-2 py-1.5">
            <Skeleton className="size-8 rounded-md" />
            <Skeleton className="h-4 w-24" />
          </div>
        </SidebarMenuItem>
      </SidebarMenu>
    );
  }

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <LeagueMonogram league={currentLeague} size={32} />
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-medium">{currentLeague.name}</span>
              </div>
              <ChevronsUpDownIcon className="ml-auto size-4" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
            align="start"
            sideOffset={4}
          >
            {otherLeagues.map((league) => (
              <DropdownMenuItem
                key={league.id}
                onClick={() => router.push(`/${league.id}`)}
                className="gap-2"
              >
                <LeagueMonogram league={league} size={24} />
                <span className="truncate">{league.name}</span>
              </DropdownMenuItem>
            ))}
            {otherLeagues.length > 0 && <DropdownMenuSeparator />}
            <DropdownMenuItem asChild>
              <Link href="/" className="gap-2">
                <LayoutGridIcon className="size-4" />
                All Leagues
              </Link>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
