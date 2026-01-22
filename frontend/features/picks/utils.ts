import type { Tournament } from "@/features/tournaments/types";

export type PickWindowState = "not_open" | "open" | "closed";

export function getPickWindowState(tournament: {
  pick_window_opens_at?: string;
  pick_window_closes_at?: string;
}): PickWindowState {
  const now = new Date();
  const opensAt = tournament.pick_window_opens_at
    ? new Date(tournament.pick_window_opens_at)
    : null;
  const closesAt = tournament.pick_window_closes_at
    ? new Date(tournament.pick_window_closes_at)
    : null;

  if (!opensAt || !closesAt) return "closed";
  if (now < opensAt) return "not_open";
  if (now >= opensAt && now < closesAt) return "open";
  return "closed";
}

export function formatPickWindowDate(isoString: string): string {
  const date = new Date(isoString);
  return date.toLocaleDateString("en-US", {
    weekday: "short",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

export function formatCountdown(isoString: string): string {
  const target = new Date(isoString);
  const now = new Date();
  const diffMs = target.getTime() - now.getTime();

  if (diffMs <= 0) return "";

  const diffMinutes = Math.floor(diffMs / (1000 * 60));
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  if (diffDays >= 1) {
    const remainingHours = diffHours % 24;
    return remainingHours > 0 ? `${diffDays}d ${remainingHours}h` : `${diffDays}d`;
  }

  if (diffHours >= 1) {
    const remainingMinutes = diffMinutes % 60;
    return remainingMinutes > 0 ? `${diffHours}h ${remainingMinutes}m` : `${diffHours}h`;
  }

  return `${diffMinutes}m`;
}

export function getPickWindowCountdown(tournament: Tournament): {
  label: string;
  countdown: string;
  state: PickWindowState;
} {
  const state = getPickWindowState(tournament);

  switch (state) {
    case "not_open":
      return {
        label: "Opens in",
        countdown: formatCountdown(tournament.pick_window_opens_at!),
        state,
      };
    case "open":
      return {
        label: "Closes in",
        countdown: formatCountdown(tournament.pick_window_closes_at!),
        state,
      };
    case "closed":
    default:
      return { label: "", countdown: "", state };
  }
}
