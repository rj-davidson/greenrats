import {
  formatPositionDisplay,
  buildPositionCounts,
  formatRankDisplay,
  buildRankCounts,
  calculateRanks,
  sumEarnings,
  aggregateByUser,
  formatCurrency,
} from "./index";
import { describe, expect, it } from "vitest";

describe("formatPositionDisplay", () => {
  it("returns CUT for cut status", () => {
    expect(formatPositionDisplay(5, "cut")).toBe("CUT");
  });

  it("returns WD for withdrawn status", () => {
    expect(formatPositionDisplay(5, "withdrawn")).toBe("WD");
  });

  it("returns dash for null position", () => {
    expect(formatPositionDisplay(null, "active")).toBe("-");
  });

  it("returns dash for zero position", () => {
    expect(formatPositionDisplay(0, "active")).toBe("-");
  });

  it("returns dash for negative position", () => {
    expect(formatPositionDisplay(-1, "active")).toBe("-");
  });

  it("returns position without T prefix when not tied", () => {
    const counts = new Map([[5, 1]]);
    expect(formatPositionDisplay(5, "active", counts)).toBe("5");
  });

  it("returns position with T prefix when tied", () => {
    const counts = new Map([[5, 2]]);
    expect(formatPositionDisplay(5, "active", counts)).toBe("T5");
  });

  it("returns position without T prefix when no counts provided", () => {
    expect(formatPositionDisplay(5, "active")).toBe("5");
  });
});

describe("buildPositionCounts", () => {
  it("counts positions for active entries", () => {
    const entries = [
      { position: 1, status: "active" },
      { position: 2, status: "active" },
      { position: 2, status: "active" },
      { position: 4, status: "active" },
    ];
    const counts = buildPositionCounts(entries);
    expect(counts.get(1)).toBe(1);
    expect(counts.get(2)).toBe(2);
    expect(counts.get(4)).toBe(1);
  });

  it("excludes cut entries from counts", () => {
    const entries = [
      { position: 1, status: "active" },
      { position: 1, status: "cut" },
    ];
    const counts = buildPositionCounts(entries);
    expect(counts.get(1)).toBe(1);
  });

  it("excludes withdrawn entries from counts", () => {
    const entries = [
      { position: 1, status: "active" },
      { position: 1, status: "withdrawn" },
    ];
    const counts = buildPositionCounts(entries);
    expect(counts.get(1)).toBe(1);
  });

  it("excludes zero/null positions from counts", () => {
    const entries = [
      { position: 0, status: "active" },
      { position: null, status: "active" },
      { position: 1, status: "active" },
    ];
    const counts = buildPositionCounts(entries);
    expect(counts.get(0)).toBeUndefined();
    expect(counts.get(1)).toBe(1);
  });
});

describe("formatRankDisplay", () => {
  it("returns rank without T prefix when not tied", () => {
    const counts = new Map([[1, 1]]);
    expect(formatRankDisplay(1, counts)).toBe("1");
  });

  it("returns rank with T prefix when tied", () => {
    const counts = new Map([[1, 2]]);
    expect(formatRankDisplay(1, counts)).toBe("T1");
  });

  it("returns rank without T prefix when count not found", () => {
    const counts = new Map<number, number>();
    expect(formatRankDisplay(1, counts)).toBe("1");
  });
});

describe("buildRankCounts", () => {
  it("counts ranks based on earnings ties", () => {
    const entries = [{ earnings: 100000 }, { earnings: 100000 }, { earnings: 50000 }];
    const counts = buildRankCounts(entries);
    expect(counts.get(1)).toBe(2);
    expect(counts.get(3)).toBe(1);
  });

  it("handles single entry", () => {
    const entries = [{ earnings: 100000 }];
    const counts = buildRankCounts(entries);
    expect(counts.get(1)).toBe(1);
  });

  it("handles empty array", () => {
    const counts = buildRankCounts([]);
    expect(counts.size).toBe(0);
  });
});

describe("calculateRanks", () => {
  it("assigns ranks based on earnings", () => {
    const entries = [
      { earnings: 100000, name: "A" },
      { earnings: 50000, name: "B" },
      { earnings: 25000, name: "C" },
    ];
    const ranked = calculateRanks(entries);
    expect(ranked[0].rank).toBe(1);
    expect(ranked[1].rank).toBe(2);
    expect(ranked[2].rank).toBe(3);
  });

  it("assigns same rank to tied earnings", () => {
    const entries = [
      { earnings: 100000, name: "A" },
      { earnings: 100000, name: "B" },
      { earnings: 50000, name: "C" },
    ];
    const ranked = calculateRanks(entries);
    expect(ranked[0].rank).toBe(1);
    expect(ranked[1].rank).toBe(1);
    expect(ranked[2].rank).toBe(3);
  });

  it("preserves original entry properties", () => {
    const entries = [{ earnings: 100000, name: "Test" }];
    const ranked = calculateRanks(entries);
    expect(ranked[0].name).toBe("Test");
    expect(ranked[0].earnings).toBe(100000);
  });
});

describe("sumEarnings", () => {
  it("sums earnings from all items", () => {
    const items = [{ earnings: 100000 }, { earnings: 50000 }, { earnings: 25000 }];
    expect(sumEarnings(items)).toBe(175000);
  });

  it("returns 0 for empty array", () => {
    expect(sumEarnings([])).toBe(0);
  });

  it("handles single item", () => {
    expect(sumEarnings([{ earnings: 100000 }])).toBe(100000);
  });
});

describe("aggregateByUser", () => {
  it("aggregates earnings and pick count by user", () => {
    const picks = [
      { user_id: "user1", earnings: 100000 },
      { user_id: "user1", earnings: 50000 },
      { user_id: "user2", earnings: 75000 },
    ];
    const result = aggregateByUser(picks);
    expect(result.get("user1")).toEqual({ earnings: 150000, pickCount: 2 });
    expect(result.get("user2")).toEqual({ earnings: 75000, pickCount: 1 });
  });

  it("returns empty map for empty array", () => {
    const result = aggregateByUser([]);
    expect(result.size).toBe(0);
  });
});

describe("formatCurrency", () => {
  it("formats positive amounts", () => {
    expect(formatCurrency(100000)).toBe("$100,000");
  });

  it("formats zero", () => {
    expect(formatCurrency(0)).toBe("$0");
  });

  it("formats negative amounts", () => {
    expect(formatCurrency(-50000)).toBe("-$50,000");
  });

  it("rounds to whole numbers", () => {
    expect(formatCurrency(100000.75)).toBe("$100,001");
  });
});
