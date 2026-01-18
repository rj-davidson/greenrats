"use client";

import { CreateLeagueDialog } from "./CreateLeagueDialog";
import { LeaguesList } from "./LeaguesList";

export function LeaguesSection() {
  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-semibold">Your Leagues</h2>
        <CreateLeagueDialog />
      </div>
      <LeaguesList />
    </section>
  );
}
