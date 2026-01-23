"use client";

import { UsersTable } from "@/features/admin/components";

export default function AdminUsersPage() {
  return (
    <div>
      <h2 className="mb-4 text-xl font-semibold">Users</h2>
      <UsersTable />
    </div>
  );
}
