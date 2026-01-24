"use client";

import { Avatar, AvatarFallback } from "@/components/shadcn/avatar";
import { Button } from "@/components/shadcn/button";
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
import { Skeleton } from "@/components/shadcn/skeleton";
import { useCurrentUser } from "@/features/users/queries";
import { signOut } from "@workos-inc/authkit-nextjs";
import {
  BookOpenIcon,
  LogOutIcon,
  MailIcon,
  MonitorIcon,
  MoonIcon,
  RatIcon,
  SunIcon,
} from "lucide-react";
import { useTheme } from "next-themes";
import Link from "next/link";

function UserMenu() {
  const { data: user, isLoading, isError } = useCurrentUser();
  const { setTheme, theme } = useTheme();

  if (isLoading) {
    return <Skeleton className="size-8 rounded-full" />;
  }

  if (isError || !user) {
    return (
      <Button asChild size="sm">
        <Link href="/login">Sign in</Link>
      </Button>
    );
  }

  const displayName = user.display_name || user.email || "User";
  const initial = displayName.charAt(0).toUpperCase();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="rounded-full">
          <Avatar className="size-8">
            <AvatarFallback>{initial}</AvatarFallback>
          </Avatar>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-56">
        <div className="px-2 py-1.5">
          <p className="text-sm font-medium">{displayName}</p>
          {user.email && user.email !== displayName && (
            <p className="text-xs text-muted-foreground">{user.email}</p>
          )}
        </div>
        <DropdownMenuSeparator />
        <DropdownMenuSub>
          <DropdownMenuSubTrigger>
            {theme === "dark" ? (
              <MoonIcon className="mr-2 size-4" />
            ) : theme === "light" ? (
              <SunIcon className="mr-2 size-4" />
            ) : (
              <MonitorIcon className="mr-2 size-4" />
            )}
            Theme
          </DropdownMenuSubTrigger>
          <DropdownMenuSubContent>
            <DropdownMenuItem onClick={() => setTheme("light")}>
              <SunIcon className="mr-2 size-4" />
              Light
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => setTheme("dark")}>
              <MoonIcon className="mr-2 size-4" />
              Dark
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => setTheme("system")}>
              <MonitorIcon className="mr-2 size-4" />
              System
            </DropdownMenuItem>
          </DropdownMenuSubContent>
        </DropdownMenuSub>
        <DropdownMenuItem asChild>
          <a href="mailto:dev@greenrats.com">
            <MailIcon className="mr-2 size-4" />
            Contact Support
          </a>
        </DropdownMenuItem>
        <DropdownMenuItem asChild>
          <Link href="/rules">
            <BookOpenIcon className="mr-2 size-4" />
            Rules
          </Link>
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={async () => {
            await signOut();
          }}
        >
          <LogOutIcon className="mr-2 size-4" />
          Sign out
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export function TopBar() {
  return (
    <header className="flex h-16 items-center justify-between border-b px-4">
      <Link href="/" className="flex items-center gap-2 font-serif text-2xl tracking-wide">
        greenrats
        <RatIcon className="size-6 text-primary" />
      </Link>
      <UserMenu />
    </header>
  );
}
