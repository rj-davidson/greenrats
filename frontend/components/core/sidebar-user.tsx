"use client";

import { Avatar, AvatarFallback } from "@/components/shadcn/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "@/components/shadcn/dropdown-menu";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/shadcn/sidebar";
import { Skeleton } from "@/components/shadcn/skeleton";
import { useCurrentUser } from "@/features/users/queries";
import { signOut } from "@workos-inc/authkit-nextjs";
import {
  BookOpenIcon,
  ChevronsUpDownIcon,
  LogOutIcon,
  MailIcon,
  MonitorIcon,
  MoonIcon,
  SunIcon,
} from "lucide-react";
import { useTheme } from "next-themes";
import Link from "next/link";

export function SidebarUser() {
  const { data: user, isLoading } = useCurrentUser();
  const { isMobile } = useSidebar();
  const { setTheme, theme } = useTheme();

  if (isLoading) {
    return (
      <SidebarMenu>
        <SidebarMenuItem>
          <div className="flex items-center gap-2 px-2 py-1.5">
            <Skeleton className="size-8 rounded-lg" />
            <Skeleton className="h-4 w-24" />
          </div>
        </SidebarMenuItem>
      </SidebarMenu>
    );
  }

  const displayName = user?.display_name || user?.email || "User";
  const initial = displayName.charAt(0).toUpperCase();

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <Avatar className="size-8 rounded-lg">
                <AvatarFallback className="rounded-lg">{initial}</AvatarFallback>
              </Avatar>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-medium">{displayName}</span>
              </div>
              <ChevronsUpDownIcon className="ml-auto size-4" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
            side={isMobile ? "bottom" : "right"}
            align="end"
            sideOffset={4}
          >
            <DropdownMenuSub>
              <DropdownMenuSubTrigger>
                {theme === "dark" ? (
                  <MoonIcon />
                ) : theme === "light" ? (
                  <SunIcon />
                ) : (
                  <MonitorIcon />
                )}
                Theme
              </DropdownMenuSubTrigger>
              <DropdownMenuSubContent>
                <DropdownMenuItem onClick={() => setTheme("light")}>
                  <SunIcon />
                  Light
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => setTheme("dark")}>
                  <MoonIcon />
                  Dark
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => setTheme("system")}>
                  <MonitorIcon />
                  System
                </DropdownMenuItem>
              </DropdownMenuSubContent>
            </DropdownMenuSub>
            <DropdownMenuItem asChild>
              <a href="mailto:dev@greenrats.com">
                <MailIcon />
                Contact Support
              </a>
            </DropdownMenuItem>
            <DropdownMenuItem asChild>
              <Link href="/rules">
                <BookOpenIcon />
                Rules
              </Link>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={async () => {
                await signOut();
              }}
            >
              <LogOutIcon />
              Sign out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
