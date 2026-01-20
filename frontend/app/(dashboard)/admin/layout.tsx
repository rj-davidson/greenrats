import { redirect } from "next/navigation";
import { makeServerRequest } from "@/lib/query/server-requestor";
import type { User } from "@/features/users/types";
import { AdminNav } from "@/features/admin/components";

async function getUser(): Promise<User | null> {
  try {
    return await makeServerRequest.get<User>("/api/v1/users/me");
  } catch {
    return null;
  }
}

export default async function AdminLayout({ children }: { children: React.ReactNode }) {
  const user = await getUser();

  if (!user?.is_admin) {
    redirect("/");
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Admin Dashboard</h1>
        <p className="text-muted-foreground">Manage users, leagues, and automations</p>
      </div>
      <AdminNav />
      {children}
    </div>
  );
}
