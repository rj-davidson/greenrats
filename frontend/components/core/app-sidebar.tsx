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
  return <span className="size-2 animate-pulse rounded-full bg-primary" />;
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
  const { data: upcomingData, isLoading: upcomingLoading } = useTournaments({
    status: "upcoming",
    limit: 1,
  });
  const { data: completedData, isLoading: completedLoading } = useTournaments({
    status: "completed",
    limit: 1,
  });

  const isLoading = activeLoading || upcomingLoading || completedLoading;

  const currentTournament = activeData?.tournament ?? upcomingData?.tournaments[0] ?? null;
  const recentCompleted = completedData?.tournaments[0] ?? null;

  return { currentTournament, recentCompleted, isLoading };
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
  const { currentTournament, recentCompleted } = useSidebarTournaments();
  const { leagues, hasMore } = useSidebarLeagues();
  const { data: user } = useCurrentUser();

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem className="w-full flex items-center justify-center">
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
        {(currentTournament || recentCompleted) && (
          <SidebarGroup>
            <SidebarGroupLabel>Upcoming</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {currentTournament && (
                  <SidebarMenuItem>
                    <SidebarMenuButton
                      asChild
                      isActive={pathname === `/tournaments/${currentTournament.id}`}
                      tooltip={currentTournament.name}
                    >
                      <Link href={`/tournaments/${currentTournament.id}`}>
                        <TournamentIcon status={currentTournament.status} />
                        <span className="truncate">{currentTournament.name}</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                )}
                {recentCompleted && recentCompleted.id !== currentTournament?.id && (
                  <SidebarMenuItem>
                    <SidebarMenuButton
                      asChild
                      isActive={pathname === `/tournaments/${recentCompleted.id}`}
                      tooltip={recentCompleted.name}
                    >
                      <Link href={`/tournaments/${recentCompleted.id}`}>
                        <TournamentIcon status={recentCompleted.status} />
                        <span className="truncate">{recentCompleted.name}</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                )}
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
