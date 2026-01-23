"use client";

import { LeaguesTable } from "@/features/admin/components";

export default function AdminLeaguesPage() {
  return (
    <div>
      <h2 className="mb-4 text-xl font-semibold">Leagues</h2>
      <LeaguesTable />
    </div>
  );
}
