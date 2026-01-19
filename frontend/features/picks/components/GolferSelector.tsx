"use client";

import type { AvailableGolfer } from "../types";
import { GolferCard } from "./GolferCard";
import { Input } from "@/components/shadcn/input";
import { Skeleton } from "@/components/shadcn/skeleton";
import { SearchIcon } from "lucide-react";
import { useMemo, useState } from "react";

interface GolferSelectorProps {
  golfers: AvailableGolfer[];
  selectedGolferId?: string;
  onSelect: (golfer: AvailableGolfer) => void;
  isLoading?: boolean;
}

export function GolferSelector({
  golfers,
  selectedGolferId,
  onSelect,
  isLoading,
}: GolferSelectorProps) {
  const [search, setSearch] = useState("");

  const filteredGolfers = useMemo(() => {
    if (!search.trim()) return golfers;
    const query = search.toLowerCase();
    return golfers.filter(
      (g) =>
        g.name.toLowerCase().includes(query) ||
        g.country?.toLowerCase().includes(query) ||
        g.country_code.toLowerCase().includes(query),
    );
  }, [golfers, search]);

  const sortedGolfers = useMemo(() => {
    return [...filteredGolfers].sort((a, b) => {
      if (a.owgr && b.owgr) return a.owgr - b.owgr;
      if (a.owgr) return -1;
      if (b.owgr) return 1;
      return a.name.localeCompare(b.name);
    });
  }, [filteredGolfers]);

  if (isLoading) {
    return (
      <div className="space-y-3">
        <Skeleton className="h-9 w-full" />
        {[1, 2, 3, 4, 5].map((i) => (
          <Skeleton key={i} className="h-16 w-full" />
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-3">
      <div className="relative">
        <SearchIcon className="text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2" />
        <Input
          placeholder="Search golfers..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="pl-9"
        />
      </div>
      <div className="text-muted-foreground text-sm">
        {sortedGolfers.length} golfer{sortedGolfers.length !== 1 ? "s" : ""} available
      </div>
      <div className="max-h-96 space-y-2 overflow-y-auto pr-1">
        {sortedGolfers.length === 0 ? (
          <div className="text-muted-foreground py-8 text-center">
            {search ? "No golfers match your search" : "No golfers available"}
          </div>
        ) : (
          sortedGolfers.map((golfer) => (
            <GolferCard
              key={golfer.id}
              golfer={golfer}
              selected={golfer.id === selectedGolferId}
              onClick={() => onSelect(golfer)}
            />
          ))
        )}
      </div>
    </div>
  );
}
