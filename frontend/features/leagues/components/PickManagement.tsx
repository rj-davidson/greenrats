"use client";

import { SearchableSelect } from "@/components/core/searchable-select";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/shadcn/select";
import { Skeleton } from "@/components/shadcn/skeleton";
import { useLeagueMembers, useLeagueTournaments } from "@/features/leagues/queries";
import type { LeagueMember } from "@/features/leagues/types";
import {
  useAvailableGolfersForUser,
  useCreatePickForUser,
  useOverridePick,
} from "@/features/picks/queries";
import type { AvailableGolfer } from "@/features/picks/types";
import { cn } from "@/lib/utils";
import { AlertTriangleIcon, CheckIcon, EditIcon } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

interface PickManagementProps {
  leagueId: string;
}

export function PickManagement({ leagueId }: PickManagementProps) {
  const [selectedTournamentId, setSelectedTournamentId] = useState<string | null>(null);
  const [selectedMember, setSelectedMember] = useState<LeagueMember | null>(null);
  const [selectedGolfer, setSelectedGolfer] = useState<AvailableGolfer | null>(null);
  const [confirmDialogOpen, setConfirmDialogOpen] = useState(false);

  const { data: tournamentsData, isLoading: tournamentsLoading } = useLeagueTournaments(leagueId);
  const { data: membersData, isLoading: membersLoading } = useLeagueMembers(
    leagueId,
    selectedTournamentId ?? undefined,
  );
  const { data: golfersData, isLoading: golfersLoading } = useAvailableGolfersForUser(
    leagueId,
    selectedTournamentId ?? "",
    selectedMember?.id ?? "",
  );
  const overridePick = useOverridePick();
  const createPickForUser = useCreatePickForUser();

  const pastTournaments =
    tournamentsData?.tournaments.filter((t) => t.status === "completed" || t.status === "active") ??
    [];

  const selectedTournament = pastTournaments.find((t) => t.id === selectedTournamentId);

  const handleTournamentChange = (tournamentId: string) => {
    setSelectedTournamentId(tournamentId);
    setSelectedMember(null);
    setSelectedGolfer(null);
  };

  const handleMemberSelect = (member: LeagueMember) => {
    setSelectedMember(member);
    setSelectedGolfer(null);
  };

  const handleGolferSelect = (golfer: AvailableGolfer) => {
    if (golfer.is_used) return;
    setSelectedGolfer(golfer);
  };

  const handleConfirmChange = async () => {
    if (!selectedMember || !selectedGolfer || !selectedTournamentId) return;

    try {
      if (selectedMember.pick) {
        await overridePick.mutateAsync({
          leagueId,
          pickId: selectedMember.pick.id,
          golferId: selectedGolfer.id,
          tournamentId: selectedTournamentId,
        });
        toast.success(`Changed ${selectedMember.display_name}'s pick to ${selectedGolfer.name}`);
      } else {
        await createPickForUser.mutateAsync({
          leagueId,
          userId: selectedMember.id,
          tournamentId: selectedTournamentId,
          golferId: selectedGolfer.id,
        });
        toast.success(`Added ${selectedGolfer.name} as ${selectedMember.display_name}'s pick`);
      }
      setConfirmDialogOpen(false);
      setSelectedMember(null);
      setSelectedGolfer(null);
    } catch {
      toast.error(selectedMember.pick ? "Failed to change pick" : "Failed to add pick");
    }
  };

  const isAddingPick = selectedMember && !selectedMember.pick;
  const canProceedToConfirm = selectedMember && selectedGolfer;

  return (
    <>
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center gap-2">
            <EditIcon className="size-5 text-primary" />
            <CardTitle className="text-base">Modify Picks</CardTitle>
          </div>
          <CardDescription>Change a member&apos;s pick for any tournament</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">

          <div className="space-y-2">
            <label className="text-sm font-medium">1. Select Tournament</label>
            {tournamentsLoading ? (
              <Skeleton className="h-10 w-full" />
            ) : pastTournaments.length === 0 ? (
              <p className="text-sm text-muted-foreground">No past tournaments available</p>
            ) : (
              <Select value={selectedTournamentId ?? ""} onValueChange={handleTournamentChange}>
                <SelectTrigger>
                  <SelectValue placeholder="Select a tournament..." />
                </SelectTrigger>
                <SelectContent>
                  {pastTournaments.map((tournament) => (
                    <SelectItem key={tournament.id} value={tournament.id}>
                      {tournament.name}
                      <span className="ml-2 text-xs text-muted-foreground">
                        ({tournament.status})
                      </span>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </div>

          {selectedTournamentId && (
            <div className="space-y-2">
              <label className="text-sm font-medium">2. Select Member</label>
              {membersLoading ? (
                <Skeleton className="h-10 w-full" />
              ) : !membersData?.members.length ? (
                <p className="text-sm text-muted-foreground">No members found</p>
              ) : (
                <SearchableSelect
                  items={membersData.members}
                  value={selectedMember}
                  onSelect={handleMemberSelect}
                  getKey={(member) => member.id}
                  getLabel={(member) => member.display_name}
                  placeholder="Search for a member..."
                  searchPlaceholder="Type to search..."
                  emptyMessage="No members found."
                  renderItem={(member, isSelected) => (
                    <div className="flex w-full items-center justify-between">
                      <div className="flex items-center gap-2">
                        <CheckIcon
                          className={cn("size-4", isSelected ? "opacity-100" : "opacity-0")}
                        />
                        <span>{member.display_name}</span>
                      </div>
                      {member.pick ? (
                        <span className="text-xs text-muted-foreground">
                          {member.pick.golfer_name}
                        </span>
                      ) : (
                        <span className="text-xs text-muted-foreground italic">No pick</span>
                      )}
                    </div>
                  )}
                />
              )}
            </div>
          )}

          {selectedMember && (
            <div className="space-y-2">
              <label className="text-sm font-medium">
                3. Select {selectedMember.pick ? "New " : ""}Golfer
              </label>
              {golfersLoading ? (
                <Skeleton className="h-10 w-full" />
              ) : !golfersData?.golfers.length ? (
                <p className="text-sm text-muted-foreground">No golfers available</p>
              ) : (
                <SearchableSelect
                  items={golfersData.golfers}
                  value={selectedGolfer}
                  onSelect={handleGolferSelect}
                  getKey={(golfer) => golfer.id}
                  getLabel={(golfer) => golfer.name}
                  placeholder="Search for a golfer..."
                  searchPlaceholder="Type to search..."
                  emptyMessage="No golfers found."
                  isItemDisabled={(golfer) =>
                    golfer.is_used || golfer.id === selectedMember.pick?.golfer_id
                  }
                  renderItem={(golfer, isSelected) => {
                    const isCurrentPick = golfer.id === selectedMember.pick?.golfer_id;
                    return (
                      <div className="flex w-full items-center justify-between">
                        <div className="flex items-center gap-2">
                          <CheckIcon
                            className={cn("size-4", isSelected ? "opacity-100" : "opacity-0")}
                          />
                          <span className={cn(isCurrentPick && "font-medium")}>{golfer.name}</span>
                          {isCurrentPick && (
                            <span className="text-xs text-muted-foreground">(current)</span>
                          )}
                        </div>
                        <div className="flex items-center gap-2 text-xs text-muted-foreground">
                          {golfer.is_used && <span>Used: {golfer.used_for_tournament}</span>}
                          {golfer.owgr && <span>OWGR: {golfer.owgr}</span>}
                        </div>
                      </div>
                    );
                  }}
                />
              )}
            </div>
          )}

          {canProceedToConfirm && (
            <Button className="w-full" onClick={() => setConfirmDialogOpen(true)}>
              {isAddingPick ? "Add Pick" : "Change Pick"}
            </Button>
          )}
        </CardContent>
      </Card>

      <Dialog open={confirmDialogOpen} onOpenChange={setConfirmDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Confirm Pick {isAddingPick ? "Addition" : "Change"}</DialogTitle>
            <DialogDescription>
              Are you sure you want to {isAddingPick ? "add" : "change"} this pick?
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-3 py-4">
            <div className="rounded-lg bg-muted p-4">
              <p className="text-sm">
                <span className="font-medium">Tournament:</span> {selectedTournament?.name}
              </p>
              <p className="text-sm">
                <span className="font-medium">Member:</span> {selectedMember?.display_name}
              </p>
              {selectedMember?.pick && (
                <p className="text-sm">
                  <span className="font-medium">Current Pick:</span>{" "}
                  {selectedMember.pick.golfer_name}
                </p>
              )}
              <p className="text-sm">
                <span className="font-medium">{isAddingPick ? "Pick" : "New Pick"}:</span>{" "}
                {selectedGolfer?.name}
              </p>
            </div>
            {selectedTournament?.status === "completed" && (
              <div className="flex items-start gap-2 rounded-lg border border-amber-500/50 bg-amber-50 p-3 dark:bg-amber-950/20">
                <AlertTriangleIcon className="mt-0.5 size-4 shrink-0 text-amber-600" />
                <p className="text-sm text-amber-700 dark:text-amber-400">
                  This tournament has already completed. {isAddingPick ? "Adding" : "Changing"} this
                  pick will affect the leaderboard and earnings calculations.
                </p>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setConfirmDialogOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleConfirmChange}
              disabled={overridePick.isPending || createPickForUser.isPending}
            >
              {overridePick.isPending || createPickForUser.isPending
                ? isAddingPick
                  ? "Adding..."
                  : "Changing..."
                : isAddingPick
                  ? "Confirm Add"
                  : "Confirm Change"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
