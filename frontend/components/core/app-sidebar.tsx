"use client";

import { SidebarUser } from "@/components/core/sidebar-user";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/shadcn/sidebar";
import { LeagueMonogram } from "@/features/leagues/components";
import { useUserLeagues } from "@/features/leagues/queries";
import { useActiveTournament, useTournaments } from "@/features/tournaments/queries";
import type { Tournament } from "@/features/tournaments/types";
import { useCurrentUser } from "@/features/users/queries";
import { CalendarIcon, ChevronRightIcon, SettingsIcon, TrophyIcon } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useMemo } from "react";

const MAX_SIDEBAR_LEAGUES = 6;

function LiveDot() {
  return (
    <span className="flex size-4 items-center justify-center">
      <span className="size-2 animate-pulse rounded-full bg-primary" />
    </span>
  );
}

function TournamentIcon({ status }: { status: Tournament["status"] }) {
  switch (status) {
    case "active":
      return <LiveDot />;
    case "upcoming":
      return <CalendarIcon className="size-4" />;
    case "completed":
      return <TrophyIcon className="size-4" />;
  }
}

function useSidebarTournaments() {
  const { data: activeData, isLoading: activeLoading } = useActiveTournament();
  const { data: completedData, isLoading: completedLoading } = useTournaments({
    status: "completed",
    limit: 1,
  });
  const { data: upcomingData, isLoading: upcomingLoading } = useTournaments({
    status: "upcoming",
    limit: 2,
  });

  const isLoading = activeLoading || completedLoading || upcomingLoading;

  const activeTournament = activeData?.tournament ?? null;
  const recentCompleted = completedData?.tournaments[0] ?? null;
  const upcomingTournaments = upcomingData?.tournaments ?? [];

  const tournaments: Tournament[] = [];

  if (recentCompleted) {
    tournaments.push(recentCompleted);
  }

  if (activeTournament) {
    tournaments.push(activeTournament);
  }

  for (const upcoming of upcomingTournaments) {
    if (tournaments.length >= 3) break;
    tournaments.push(upcoming);
  }

  return { tournaments, isLoading };
}

function useSidebarLeagues() {
  const { data, isLoading } = useUserLeagues();

  const sortedLeagues = useMemo(() => {
    if (!data?.leagues) return [];
    return [...data.leagues].sort((a, b) => a.name.localeCompare(b.name));
  }, [data]);

  const displayedLeagues = sortedLeagues.slice(0, MAX_SIDEBAR_LEAGUES);
  const hasMore = sortedLeagues.length > MAX_SIDEBAR_LEAGUES;

  return { leagues: displayedLeagues, hasMore, isLoading };
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const pathname = usePathname();
  const { tournaments } = useSidebarTournaments();
  const { leagues, hasMore } = useSidebarLeagues();
  const { data: user } = useCurrentUser();

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem className="flex w-full items-center justify-center">
            <Link href="/">
              <Image
                src="/assets/logo.png"
                alt="GreenRats"
                width={220}
                height={83}
                className="group-data-[collapsible=icon]:hidden"
              />
              <Image
                src="/assets/logo_square.png"
                alt="GreenRats"
                width={32}
                height={32}
                className="hidden group-data-[collapsible=icon]:block"
              />
            </Link>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        {tournaments.length > 0 && (
          <SidebarGroup>
            <SidebarGroupLabel>Tournaments</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {tournaments.map((tournament) => (
                  <SidebarMenuItem key={tournament.id}>
                    <SidebarMenuButton
                      asChild
                      isActive={pathname === `/tournaments/${tournament.id}`}
                      tooltip={tournament.name}
                    >
                      <Link href={`/tournaments/${tournament.id}`}>
                        <TournamentIcon status={tournament.status} />
                        <span className="truncate">{tournament.name}</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        )}
        {leagues.length > 0 && (
          <SidebarGroup>
            <SidebarGroupLabel>Leagues</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {leagues.map((league) => (
                  <SidebarMenuItem key={league.id}>
                    <SidebarMenuButton
                      asChild
                      isActive={pathname === `/leagues/${league.id}`}
                      tooltip={league.name}
                    >
                      <Link href={`/leagues/${league.id}`}>
                        <LeagueMonogram league={league} size={16} />
                        <span className="truncate">{league.name}</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
                {hasMore && (
                  <SidebarMenuItem>
                    <SidebarMenuButton asChild tooltip="View all leagues">
                      <Link href="/">
                        <ChevronRightIcon />
                        <span>View more</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                )}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        )}
        {user?.is_admin && (
          <SidebarGroup>
            <SidebarGroupLabel>Admin</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                <SidebarMenuItem>
                  <SidebarMenuButton
                    asChild
                    isActive={pathname.startsWith("/admin")}
                    tooltip="Admin Dashboard"
                  >
                    <Link href="/admin">
                      <SettingsIcon className="size-4" />
                      <span>Dashboard</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        )}
      </SidebarContent>
      <SidebarFooter>
        <SidebarUser />
      </SidebarFooter>
    </Sidebar>
  );
}
