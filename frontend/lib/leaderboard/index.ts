export function formatPositionDisplay(
  position: number | null,
  status: string,
  positionCounts?: Map<number, number>,
): string {
  if (status === "cut") return "CUT";
  if (status === "withdrawn") return "WD";
  if (!position || position <= 0) return "-";

  const isTied = (positionCounts?.get(position) ?? 0) > 1;
  return isTied ? `T${position}` : `${position}`;
}

export function buildPositionCounts<T extends { position: number | null; status: string }>(
  entries: T[],
): Map<number, number> {
  const counts = new Map<number, number>();
  for (const entry of entries) {
    if (
      entry.position &&
      entry.position > 0 &&
      entry.status !== "cut" &&
      entry.status !== "withdrawn"
    ) {
      counts.set(entry.position, (counts.get(entry.position) ?? 0) + 1);
    }
  }
  return counts;
}

export function formatRankDisplay(rank: number, rankCounts: Map<number, number>): string {
  const isTied = (rankCounts.get(rank) ?? 0) > 1;
  return isTied ? `T${rank}` : `${rank}`;
}

export function buildRankCounts<T extends { earnings: number }>(
  sortedEntries: T[],
): Map<number, number> {
  const counts = new Map<number, number>();
  let currentRank = 1;
  let prevEarnings: number | null = null;

  for (let i = 0; i < sortedEntries.length; i++) {
    if (prevEarnings !== null && sortedEntries[i].earnings < prevEarnings) {
      currentRank = i + 1;
    }
    counts.set(currentRank, (counts.get(currentRank) ?? 0) + 1);
    prevEarnings = sortedEntries[i].earnings;
  }
  return counts;
}

export function calculateRanks<T extends { earnings: number }>(
  sortedEntries: T[],
): Array<T & { rank: number }> {
  let currentRank = 1;
  let prevEarnings: number | null = null;

  return sortedEntries.map((entry, i) => {
    if (prevEarnings !== null && entry.earnings < prevEarnings) {
      currentRank = i + 1;
    }
    prevEarnings = entry.earnings;
    return { ...entry, rank: currentRank };
  });
}

export function sumEarnings<T extends { earnings: number }>(items: T[]): number {
  return items.reduce((sum, item) => sum + item.earnings, 0);
}

export function aggregateByUser<T extends { user_id: string; earnings: number }>(
  picks: T[],
): Map<string, { earnings: number; pickCount: number }> {
  const byUser = new Map<string, { earnings: number; pickCount: number }>();

  for (const pick of picks) {
    const current = byUser.get(pick.user_id) ?? { earnings: 0, pickCount: 0 };
    byUser.set(pick.user_id, {
      earnings: current.earnings + pick.earnings,
      pickCount: current.pickCount + 1,
    });
  }
  return byUser;
}

const currencyFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "USD",
  maximumFractionDigits: 0,
});

export function formatCurrency(amount: number): string {
  return currencyFormatter.format(amount);
}
