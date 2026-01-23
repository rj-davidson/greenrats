"use client";

import { Input } from "@/components/shadcn/input";
import { Skeleton } from "@/components/shadcn/skeleton";
import { Table, TableBody, TableHead, TableHeader, TableRow } from "@/components/shadcn/table";
import { GolferTableRow } from "@/features/picks/components/GolferTableRow";
import type { AvailableGolfer } from "@/features/picks/types";
import { SearchIcon } from "lucide-react";
import { useMemo, useState } from "react";

interface GolferSelectorProps {
  golfers: AvailableGolfer[];
  onSelect: (golfer: AvailableGolfer) => void;
  isLoading?: boolean;
}

export function GolferSelector({ golfers, onSelect, isLoading }: GolferSelectorProps) {
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
      const aUsed = a.is_used ?? false;
      const bUsed = b.is_used ?? false;
      if (aUsed !== bUsed) return aUsed ? 1 : -1;
      if (a.owgr && b.owgr) return a.owgr - b.owgr;
      if (a.owgr) return -1;
      if (b.owgr) return 1;
      return a.name.localeCompare(b.name);
    });
  }, [filteredGolfers]);

  const availableCount = useMemo(() => {
    return golfers.filter((g) => !g.is_used).length;
  }, [golfers]);

  if (isLoading) {
    return (
      <div className="space-y-3">
        <Skeleton className="h-9 w-full" />
        <Skeleton className="h-6 w-48" />
        <div className="space-y-1">
          {[1, 2, 3, 4, 5].map((i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      <div className="relative">
        <SearchIcon className="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search golfers..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="pl-9"
        />
      </div>
      <div className="text-sm text-muted-foreground">
        {availableCount} of {golfers.length} golfer{golfers.length !== 1 ? "s" : ""} available
      </div>
      <div className="max-h-96 overflow-y-auto">
        {sortedGolfers.length === 0 ? (
          <div className="py-8 text-center text-muted-foreground">
            {search ? "No golfers match your search" : "No golfers available"}
          </div>
        ) : (
          <Table>
            <TableHeader className="sticky top-0 bg-background">
              <TableRow>
                <TableHead>Golfer</TableHead>
                <TableHead className="hidden md:table-cell">Country</TableHead>
                <TableHead className="text-right">OWGR</TableHead>
                <TableHead className="text-right">Action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {sortedGolfers.map((golfer) => (
                <GolferTableRow key={golfer.id} golfer={golfer} onSelect={onSelect} />
              ))}
            </TableBody>
          </Table>
        )}
      </div>
    </div>
  );
}
