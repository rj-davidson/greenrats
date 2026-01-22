"use client";

import { useState } from "react";
import {
  useAdminTournaments,
  useAdminTournamentField,
  useDeleteFieldEntry,
  useTriggerSyncField,
} from "@/features/admin/queries";
import { Button } from "@/components/shadcn/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/shadcn/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/shadcn/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/shadcn/table";
import { Badge } from "@/components/shadcn/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogClose,
} from "@/components/shadcn/dialog";
import { toast } from "sonner";
import { RefreshCwIcon, Trash2Icon } from "lucide-react";
import type { FieldEntry } from "@/features/admin/types";

function getStatusBadgeVariant(status: string): "default" | "secondary" | "destructive" | "outline" {
  switch (status) {
    case "confirmed":
      return "default";
    case "alternate":
      return "secondary";
    case "withdrawn":
      return "destructive";
    case "pending":
      return "outline";
    default:
      return "default";
  }
}

function FieldEntryRow({ entry, onDelete }: { entry: FieldEntry; onDelete: () => void }) {
  return (
    <TableRow>
      <TableCell className="font-medium">
        {entry.golfer_name}
        {entry.is_amateur && <Badge variant="outline" className="ml-2">Amateur</Badge>}
      </TableCell>
      <TableCell>{entry.country_code}</TableCell>
      <TableCell>
        <Badge variant={getStatusBadgeVariant(entry.entry_status)}>
          {entry.entry_status}
        </Badge>
      </TableCell>
      <TableCell>{entry.qualifier || "-"}</TableCell>
      <TableCell>{entry.owgr_at_entry ?? "-"}</TableCell>
      <TableCell>
        <Dialog>
          <DialogTrigger asChild>
            <Button variant="ghost" size="icon">
              <Trash2Icon className="h-4 w-4" />
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Remove from Field</DialogTitle>
              <DialogDescription>
                Are you sure you want to remove {entry.golfer_name} from the tournament field?
                This action cannot be undone.
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <DialogClose asChild>
                <Button variant="outline">Cancel</Button>
              </DialogClose>
              <DialogClose asChild>
                <Button variant="destructive" onClick={onDelete}>Remove</Button>
              </DialogClose>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </TableCell>
    </TableRow>
  );
}

export function FieldsPage() {
  const [selectedTournament, setSelectedTournament] = useState<string>("");

  const { data: tournamentsData, isLoading: loadingTournaments } = useAdminTournaments();
  const { data: fieldData, isLoading: loadingField } = useAdminTournamentField(selectedTournament);
  const deleteEntry = useDeleteFieldEntry(selectedTournament);
  const syncField = useTriggerSyncField();

  const handleDeleteEntry = async (entryId: string) => {
    try {
      await deleteEntry.mutateAsync(entryId);
      toast.success("Golfer removed from field");
    } catch {
      toast.error("Failed to remove golfer");
    }
  };

  const handleSyncField = async () => {
    if (!selectedTournament) {
      toast.error("Please select a tournament");
      return;
    }
    try {
      await syncField.mutateAsync(selectedTournament);
      toast.success("Field sync started");
    } catch {
      toast.error("Failed to start field sync");
    }
  };

  const selectedTournamentName = tournamentsData?.tournaments.find(
    t => t.id === selectedTournament
  )?.name;

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Tournament Field Management</CardTitle>
          <CardDescription>
            View and manage tournament field entries
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex gap-4">
            <Select value={selectedTournament} onValueChange={setSelectedTournament}>
              <SelectTrigger className="w-[300px]">
                <SelectValue placeholder={loadingTournaments ? "Loading..." : "Select a tournament"} />
              </SelectTrigger>
              <SelectContent>
                {tournamentsData?.tournaments.map((tournament) => (
                  <SelectItem key={tournament.id} value={tournament.id}>
                    {tournament.name} ({tournament.status})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Button
              onClick={handleSyncField}
              disabled={!selectedTournament || syncField.isPending}
              variant="outline"
            >
              <RefreshCwIcon className={syncField.isPending ? "animate-spin" : ""} />
              {syncField.isPending ? "Syncing..." : "Sync Field"}
            </Button>
          </div>
        </CardContent>
      </Card>

      {selectedTournament && (
        <Card>
          <CardHeader>
            <CardTitle>
              Field: {selectedTournamentName}
            </CardTitle>
            <CardDescription>
              {loadingField
                ? "Loading..."
                : `${fieldData?.total ?? 0} entries`}
            </CardDescription>
          </CardHeader>
          <CardContent>
            {loadingField ? (
              <div className="py-8 text-center text-muted-foreground">Loading field data...</div>
            ) : fieldData?.entries.length === 0 ? (
              <div className="py-8 text-center text-muted-foreground">
                No field entries found. Click &quot;Sync Field&quot; to fetch field data from PGA Tour.
              </div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Golfer</TableHead>
                    <TableHead>Country</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Qualifier</TableHead>
                    <TableHead>OWGR</TableHead>
                    <TableHead className="w-[50px]"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {fieldData?.entries.map((entry) => (
                    <FieldEntryRow
                      key={entry.id}
                      entry={entry}
                      onDelete={() => handleDeleteEntry(entry.id)}
                    />
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
