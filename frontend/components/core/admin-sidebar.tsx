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
import { BotIcon, HomeIcon, SettingsIcon, TrophyIcon, UsersIcon } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";

export function AdminSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const pathname = usePathname();

  const navItems = [
    {
      label: "Users",
      href: "/admin/users",
      icon: UsersIcon,
      isActive: pathname === "/admin/users",
    },
    {
      label: "Leagues",
      href: "/admin/leagues",
      icon: TrophyIcon,
      isActive: pathname === "/admin/leagues",
    },
    {
      label: "Automations",
      href: "/admin/automations",
      icon: BotIcon,
      isActive: pathname === "/admin/automations",
    },
  ];

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem className="flex w-full items-center justify-center">
            <Link href="/" className="font-serif text-xl tracking-wide">
              <span className="group-data-[collapsible=icon]:hidden">GREEN RATS</span>
              <span className="hidden group-data-[collapsible=icon]:block">GR</span>
            </Link>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>
            <SettingsIcon className="mr-2 size-4" />
            Admin
          </SidebarGroupLabel>
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
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton asChild tooltip="Back to Leagues">
                  <Link href="/">
                    <HomeIcon className="size-4" />
                    <span>Back to Leagues</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        <SidebarUser />
      </SidebarFooter>
    </Sidebar>
  );
}
