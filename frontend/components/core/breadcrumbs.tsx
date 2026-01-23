"use client";

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/shadcn/breadcrumb";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  createContext,
  type Dispatch,
  Fragment,
  type ReactNode,
  type SetStateAction,
  useContext,
  useMemo,
  useState,
} from "react";

const routeNames: Record<string, string> = {
  "/create": "Create League",
  "/standings": "Standings",
  "/tournaments": "Tournaments",
  "/audit": "Audit Log",
  "/manage": "Manage",
  "/admin": "Admin",
  "/admin/users": "Users",
  "/admin/leagues": "Leagues",
  "/admin/automations": "Automations",
  "/rules": "Rules",
  "/onboarding": "Onboarding",
};

type BreadcrumbItemData = {
  name: string;
  path?: string;
};

type BreadcrumbsContextValue = {
  extraCrumbs: BreadcrumbItemData[];
  setExtraCrumbs: Dispatch<SetStateAction<BreadcrumbItemData[]>>;
};

const BreadcrumbsContext = createContext<BreadcrumbsContextValue | null>(null);

export function BreadcrumbsProvider({ children }: { children: ReactNode }) {
  const [extraCrumbs, setExtraCrumbs] = useState<BreadcrumbItemData[]>([]);
  const value = useMemo(() => ({ extraCrumbs, setExtraCrumbs }), [extraCrumbs]);

  return <BreadcrumbsContext.Provider value={value}>{children}</BreadcrumbsContext.Provider>;
}

export function useBreadcrumbs() {
  const context = useContext(BreadcrumbsContext);

  if (!context) {
    throw new Error("useBreadcrumbs must be used within BreadcrumbsProvider.");
  }

  return context;
}

export function Breadcrumbs() {
  const pathname = usePathname();
  const { extraCrumbs } = useBreadcrumbs();

  const baseCrumbs = useMemo(() => {
    if (!pathname) return [];

    const segments = pathname.split("/").filter(Boolean);

    if (segments.length === 0) {
      return [];
    }

    const crumbs = segments
      .map((_segment, index) => {
        const path = `/${segments.slice(0, index + 1).join("/")}`;
        const name = routeNames[path];

        if (!name) {
          return null;
        }

        return {
          name,
          path,
        };
      })
      .filter((crumb): crumb is { name: string; path: string } => crumb !== null);

    return crumbs;
  }, [pathname]);

  const breadcrumbs = useMemo(() => {
    const merged = [...baseCrumbs, ...extraCrumbs];

    return merged.map((crumb, index) => ({
      ...crumb,
      isLast: index === merged.length - 1,
    }));
  }, [baseCrumbs, extraCrumbs]);

  if (breadcrumbs.length === 0) {
    return null;
  }

  return (
    <Breadcrumb>
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbLink asChild>
            <Link href="/">Home</Link>
          </BreadcrumbLink>
        </BreadcrumbItem>
        {breadcrumbs.map((crumb) => (
          <Fragment key={crumb.name}>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              {crumb.isLast || !crumb.path ? (
                <BreadcrumbPage>{crumb.name}</BreadcrumbPage>
              ) : (
                <BreadcrumbLink asChild>
                  <Link href={crumb.path}>{crumb.name}</Link>
                </BreadcrumbLink>
              )}
            </BreadcrumbItem>
          </Fragment>
        ))}
      </BreadcrumbList>
    </Breadcrumb>
  );
}
