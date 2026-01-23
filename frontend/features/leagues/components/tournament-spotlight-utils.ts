import type { LeagueTournament } from "@/features/leagues/types";

export function getTournamentSpotlight(tournaments: LeagueTournament[]): LeagueTournament[] {
  if (tournaments.length === 0) return [];
  if (tournaments.length <= 3) return tournaments;

  const sorted = [...tournaments].sort(
    (a, b) => new Date(a.start_date).getTime() - new Date(b.start_date).getTime(),
  );

  const today = new Date();
  today.setHours(0, 0, 0, 0);

  let anchorIndex = sorted.findIndex((t) => new Date(t.end_date) >= today);

  if (anchorIndex === -1) {
    anchorIndex = sorted.length - 1;
  }

  let start = anchorIndex - 1;
  let end = anchorIndex + 2;

  if (start < 0) {
    start = 0;
    end = Math.min(3, sorted.length);
  } else if (end > sorted.length) {
    end = sorted.length;
    start = Math.max(0, sorted.length - 3);
  }

  return sorted.slice(start, end);
}
