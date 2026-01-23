import { AdminSidebar } from "@/components/core/admin-sidebar";
import { Breadcrumbs, BreadcrumbsProvider } from "@/components/core/breadcrumbs";
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/shadcn/sidebar";
import type { User } from "@/features/users/types";
import { makeServerRequest } from "@/lib/query/server-requestor";
import { redirect } from "next/navigation";

async function getUser(): Promise<User | null> {
  try {
    return await makeServerRequest.get<User>("/api/v1/users/me");
  } catch {
    return null;
  }
}

export default async function AdminLayout({ children }: { children: React.ReactNode }) {
  const user = await getUser();

  if (!user) {
    redirect("/login");
  }

  if (!user.is_admin) {
    redirect("/");
  }

  return (
    <SidebarProvider>
      <BreadcrumbsProvider>
        <AdminSidebar />
        <SidebarInset>
          <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
            <SidebarTrigger className="-ml-1" />
            <Breadcrumbs />
          </header>
          <main className="min-w-0 flex-1 overflow-x-hidden p-4">
            <div className="space-y-6">
              <div>
                <h1 className="text-2xl font-bold">Admin Dashboard</h1>
                <p className="text-muted-foreground">Manage users, leagues, and automations</p>
              </div>
              {children}
            </div>
          </main>
        </SidebarInset>
      </BreadcrumbsProvider>
    </SidebarProvider>
  );
}
