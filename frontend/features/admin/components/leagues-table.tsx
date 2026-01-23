"use client";

import { Button } from "@/components/shadcn/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/shadcn/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/shadcn/table";
import { useAdminLeagues, useDeleteLeague } from "@/features/admin/queries";
import { TrashIcon } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

export function LeaguesTable() {
  const { data, isLoading, error } = useAdminLeagues();
  const deleteLeague = useDeleteLeague();
  const [deleteTarget, setDeleteTarget] = useState<{ id: string; name: string } | null>(null);

  if (isLoading) {
    return <div className="py-4 text-muted-foreground">Loading leagues...</div>;
  }

  if (error) {
    return <div className="py-4 text-destructive">Failed to load leagues</div>;
  }

  if (!data || data.leagues.length === 0) {
    return <div className="py-4 text-muted-foreground">No leagues found</div>;
  }

  const handleDelete = async () => {
    if (!deleteTarget) return;

    try {
      await deleteLeague.mutateAsync(deleteTarget.id);
      toast.success(`League "${deleteTarget.name}" deleted`);
      setDeleteTarget(null);
    } catch {
      toast.error("Failed to delete league");
    }
  };

  return (
    <>
      <div className="space-y-4">
        <div className="text-sm text-muted-foreground">{data.total} total leagues</div>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Season</TableHead>
              <TableHead>Members</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="w-[80px]">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.leagues.map((league) => (
              <TableRow key={league.id}>
                <TableCell className="font-medium">{league.name}</TableCell>
                <TableCell>{league.season_year}</TableCell>
                <TableCell>{league.member_count}</TableCell>
                <TableCell className="text-muted-foreground">
                  {new Date(league.created_at).toLocaleDateString()}
                </TableCell>
                <TableCell>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    onClick={() => setDeleteTarget({ id: league.id, name: league.name })}
                  >
                    <TrashIcon className="size-4 text-destructive" />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      <Dialog open={!!deleteTarget} onOpenChange={() => setDeleteTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete League</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete &quot;{deleteTarget?.name}&quot;? This action cannot
              be undone and will remove all associated data.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteTarget(null)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleteLeague.isPending}>
              {deleteLeague.isPending ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
