"use client";

import { Input } from "@/components/shadcn/input";
import { Skeleton } from "@/components/shadcn/skeleton";
import { Table, TableBody, TableHead, TableHeader, TableRow } from "@/components/shadcn/table";
import { PickConfirmDialog } from "@/features/picks/components/PickConfirmDialog";
import { PickFieldRow } from "@/features/picks/components/PickFieldRow";
import { useCreatePick, useUpdatePick } from "@/features/picks/queries";
import type { GetPickFieldResponse, PickFieldEntry } from "@/features/picks/types";
import { SearchIcon } from "lucide-react";
import { useCallback, useMemo, useState } from "react";
import { toast } from "sonner";

interface PickFieldTableProps {
  data: GetPickFieldResponse;
  leagueId: string;
}

function PickFieldSkeleton() {
  return (
    <div className="space-y-3">
      <Skeleton className="h-9 w-full" />
      <Skeleton className="h-6 w-48" />
      <div className="space-y-1">
        {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
          <Skeleton key={i} className="h-12 w-full" />
        ))}
      </div>
    </div>
  );
}

export function PickFieldTable({ data, leagueId }: PickFieldTableProps) {
  const [search, setSearch] = useState("");
  const [expandedGolferId, setExpandedGolferId] = useState<string | null>(null);
  const [selectedEntry, setSelectedEntry] = useState<PickFieldEntry | null>(null);
  const [confirmOpen, setConfirmOpen] = useState(false);

  const createPick = useCreatePick();
  const updatePick = useUpdatePick();

  const filteredEntries = useMemo(() => {
    if (!search.trim()) return data.entries;
    const query = search.toLowerCase();
    return data.entries.filter(
      (e) =>
        e.golfer_name.toLowerCase().includes(query) ||
        e.country?.toLowerCase().includes(query) ||
        e.country_code.toLowerCase().includes(query),
    );
  }, [data.entries, search]);

  const sortedEntries = useMemo(() => {
    return [...filteredEntries].sort((a, b) => {
      const aSignal = a.signal ?? -1;
      const bSignal = b.signal ?? -1;
      if (aSignal !== bSignal) return bSignal - aSignal;
      return a.golfer_name.localeCompare(b.golfer_name);
    });
  }, [filteredEntries]);

  const toggleExpand = useCallback((golferId: string) => {
    setExpandedGolferId((prev) => (prev === golferId ? null : golferId));
  }, []);

  const handleSelect = useCallback((entry: PickFieldEntry) => {
    setSelectedEntry(entry);
    setConfirmOpen(true);
  }, []);

  const handleConfirmPick = useCallback(async () => {
    if (!selectedEntry) return;

    const isChanging = !!data.current_pick_id;

    try {
      if (isChanging && data.current_pick_id) {
        await updatePick.mutateAsync({
          pickId: data.current_pick_id,
          golferId: selectedEntry.golfer_id,
          leagueId,
          tournamentId: data.tournament_id,
        });
        toast.success(`Successfully changed pick to ${selectedEntry.golfer_name}!`);
      } else {
        await createPick.mutateAsync({
          tournament_id: data.tournament_id,
          golfer_id: selectedEntry.golfer_id,
          league_id: leagueId,
        });
        toast.success(`Successfully picked ${selectedEntry.golfer_name}!`);
      }
      setConfirmOpen(false);
      setSelectedEntry(null);
    } catch {
      toast.error(`Failed to ${isChanging ? "change" : "submit"} pick. Please try again.`);
    }
  }, [selectedEntry, data.current_pick_id, data.tournament_id, leagueId, createPick, updatePick]);

  const isSubmitting = createPick.isPending || updatePick.isPending;
  const isChanging = !!data.current_pick_id;

  return (
    <>
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
          {data.available_count} of {data.total} golfer{data.total !== 1 ? "s" : ""} available
        </div>
        {sortedEntries.length === 0 ? (
          <div className="py-8 text-center text-muted-foreground">
            {search ? "No golfers match your search" : "No golfers in field"}
          </div>
        ) : (
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-0"></TableHead>
                  <TableHead>Golfer</TableHead>
                  <TableHead className="w-16">Signal</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {sortedEntries.map((entry) => (
                  <PickFieldRow
                    key={entry.golfer_id}
                    entry={entry}
                    isExpanded={expandedGolferId === entry.golfer_id}
                    isCurrentPick={entry.golfer_id === data.current_pick_golfer_id}
                    pickWindowState={data.pick_window_state}
                    onToggle={() => toggleExpand(entry.golfer_id)}
                    onSelect={() => handleSelect(entry)}
                  />
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </div>

      <PickConfirmDialog
        golfer={
          selectedEntry
            ? {
                id: selectedEntry.golfer_id,
                name: selectedEntry.golfer_name,
                country_code: selectedEntry.country_code,
                country: selectedEntry.country,
                owgr: selectedEntry.owgr ?? undefined,
                image_url: selectedEntry.image_url,
              }
            : null
        }
        tournamentName={data.tournament_name}
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        onConfirm={handleConfirmPick}
        isSubmitting={isSubmitting}
        isChanging={isChanging}
      />
    </>
  );
}

export { PickFieldSkeleton };
