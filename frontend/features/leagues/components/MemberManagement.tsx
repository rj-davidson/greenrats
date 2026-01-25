"use client";

import { Button } from "@/components/shadcn/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/shadcn/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/shadcn/dialog";
import { Skeleton } from "@/components/shadcn/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/shadcn/table";
import { useLeagueMembers, useRemoveMember } from "@/features/leagues/queries";
import { AlertTriangleIcon, TrashIcon, UsersIcon } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

interface MemberManagementProps {
  leagueId: string;
}

export function MemberManagement({ leagueId }: MemberManagementProps) {
  const { data, isLoading } = useLeagueMembers(leagueId);
  const removeMember = useRemoveMember();
  const [removeTarget, setRemoveTarget] = useState<{ id: string; name: string } | null>(null);

  const handleRemove = async () => {
    if (!removeTarget) return;

    try {
      await removeMember.mutateAsync({ leagueId, userId: removeTarget.id });
      toast.success(`${removeTarget.name} has been removed from the league`);
      setRemoveTarget(null);
    } catch {
      toast.error("Failed to remove member");
    }
  };

  return (
    <>
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center gap-2">
            <UsersIcon className="size-5 text-primary" />
            <CardTitle className="text-base">League Members</CardTitle>
          </div>
          <CardDescription>Manage members in this league</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-2">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : !data?.members.length ? (
            <p className="text-sm text-muted-foreground">No members found</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Role</TableHead>
                  <TableHead>Joined</TableHead>
                  <TableHead className="w-[80px]">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.members.map((member) => (
                  <TableRow key={member.id}>
                    <TableCell className="font-medium">{member.display_name}</TableCell>
                    <TableCell className="capitalize">{member.role}</TableCell>
                    <TableCell className="text-muted-foreground">
                      {new Date(member.joined_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell>
                      {member.role !== "owner" && (
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          onClick={() =>
                            setRemoveTarget({ id: member.id, name: member.display_name })
                          }
                        >
                          <TrashIcon className="size-4 text-destructive" />
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Dialog open={!!removeTarget} onOpenChange={() => setRemoveTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Remove Member</DialogTitle>
            <DialogDescription>
              Are you sure you want to remove {removeTarget?.name} from the league?
            </DialogDescription>
          </DialogHeader>
          <div className="flex items-start gap-2 rounded-lg border border-destructive/50 bg-destructive/10 p-3">
            <AlertTriangleIcon className="mt-0.5 size-4 shrink-0 text-destructive" />
            <p className="text-sm text-destructive">
              This will remove {removeTarget?.name} from the league. All their picks will be deleted
              and they will lose access to the league.
            </p>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRemoveTarget(null)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleRemove} disabled={removeMember.isPending}>
              {removeMember.isPending ? "Removing..." : "Remove Member"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
