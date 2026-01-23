"use client";

import { LeagueSwitcher } from "@/components/core/league-switcher";
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
import type { League } from "@/features/leagues/types";
import { useCurrentUser } from "@/features/users/queries";
import {
  CalendarIcon,
  HistoryIcon,
  LayoutDashboardIcon,
  SettingsIcon,
  TrophyIcon,
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";

interface LeagueSidebarProps extends React.ComponentProps<typeof Sidebar> {
  league: Pick<League, "id" | "name" | "role">;
}

export function LeagueSidebar({ league, ...props }: LeagueSidebarProps) {
  const pathname = usePathname();
  const { data: user } = useCurrentUser();

  const isOwner = league.role === "owner";

  const navItems = [
    {
      label: "Dashboard",
      href: `/${league.id}`,
      icon: LayoutDashboardIcon,
      isActive: pathname === `/${league.id}`,
    },
    {
      label: "Standings",
      href: `/${league.id}/standings`,
      icon: TrophyIcon,
      isActive: pathname === `/${league.id}/standings`,
    },
    {
      label: "Tournaments",
      href: `/${league.id}/tournaments`,
      icon: CalendarIcon,
      isActive: pathname.startsWith(`/${league.id}/tournaments`),
    },
    {
      label: "Audit",
      href: `/${league.id}/audit`,
      icon: HistoryIcon,
      isActive: pathname === `/${league.id}/audit`,
    },
  ];

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <LeagueSwitcher currentLeague={league} />
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {navItems.map((item) => (
                <SidebarMenuItem key={item.href}>
                  <SidebarMenuButton asChild isActive={item.isActive} tooltip={item.label}>
                    <Link href={item.href}>
                      <item.icon className="size-4" />
                      <span>{item.label}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
              {isOwner && (
                <SidebarMenuItem>
                  <SidebarMenuButton
                    asChild
                    isActive={pathname === `/${league.id}/manage`}
                    tooltip="Manage"
                  >
                    <Link href={`/${league.id}/manage`}>
                      <SettingsIcon className="size-4" />
                      <span>Manage</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              )}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
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
