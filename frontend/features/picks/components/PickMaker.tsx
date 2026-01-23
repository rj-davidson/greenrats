"use client";

import { Badge } from "@/components/shadcn/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/shadcn/card";
import { GolferSelector } from "@/features/picks/components/GolferSelector";
import { PickConfirmDialog } from "@/features/picks/components/PickConfirmDialog";
import { useAvailableGolfers, useCreatePick, useUpdatePick } from "@/features/picks/queries";
import type { AvailableGolfer, Pick } from "@/features/picks/types";
import { getPickWindowState, formatPickWindowDate } from "@/features/picks/utils";
import type { Tournament } from "@/features/tournaments/types";
import { CalendarIcon, ClockIcon, LockIcon, UserIcon } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

interface PickMakerProps {
  leagueId: string;
  tournament: Tournament;
  currentPick?: Pick;
  onPickSuccess?: () => void;
}

export function PickMaker({ leagueId, tournament, currentPick, onPickSuccess }: PickMakerProps) {
  const [selectedGolfer, setSelectedGolfer] = useState<AvailableGolfer | null>(null);
  const [confirmOpen, setConfirmOpen] = useState(false);

  const { data: golfersData, isLoading: golfersLoading } = useAvailableGolfers(
    leagueId,
    tournament.id,
  );
  const createPick = useCreatePick();
  const updatePick = useUpdatePick();

  const isChanging = !!currentPick;
  const pickWindowState = getPickWindowState(tournament);
  const isWindowOpen = pickWindowState === "open";

  const handleSelectGolfer = (golfer: AvailableGolfer) => {
    setSelectedGolfer(golfer);
    setConfirmOpen(true);
  };

  const handleConfirmPick = async () => {
    if (!selectedGolfer) return;

    try {
      if (isChanging) {
        await updatePick.mutateAsync({
          pickId: currentPick.id,
          golferId: selectedGolfer.id,
          leagueId,
          tournamentId: tournament.id,
        });
        toast.success(`Successfully changed pick to ${selectedGolfer.name}!`);
      } else {
        await createPick.mutateAsync({
          tournament_id: tournament.id,
          golfer_id: selectedGolfer.id,
          league_id: leagueId,
        });
        toast.success(`Successfully picked ${selectedGolfer.name}!`);
      }
      setConfirmOpen(false);
      setSelectedGolfer(null);
      onPickSuccess?.();
    } catch {
      toast.error(`Failed to ${isChanging ? "change" : "submit"} pick. Please try again.`);
    }
  };

  const getClosedReason = (): string => {
    if (pickWindowState === "not_open" && tournament.pick_window_opens_at) {
      return `Pick window opens ${formatPickWindowDate(tournament.pick_window_opens_at)}`;
    }
    return "Pick window is closed";
  };

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-start justify-between gap-2">
            <div>
              <CardTitle>{isChanging ? "Change Your Pick" : "Make Your Pick"}</CardTitle>
              <CardDescription>
                {isChanging
                  ? "Select a different golfer to replace your current pick"
                  : "Select your golfer for this tournament"}
              </CardDescription>
            </div>
            <Badge variant={isWindowOpen ? "default" : "secondary"}>
              {isWindowOpen ? "Open" : "Closed"}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {currentPick && (
            <div className="flex items-center gap-3 rounded-lg border bg-muted/50 p-3">
              <UserIcon className="size-5 text-muted-foreground" />
              <div>
                <div className="text-sm font-medium">Current Pick</div>
                <div className="text-sm text-muted-foreground">{currentPick.golfer_name}</div>
              </div>
            </div>
          )}

          {tournament.pick_window_opens_at && tournament.pick_window_closes_at && (
            <div className="flex flex-wrap gap-4 text-sm text-muted-foreground">
              <div className="flex items-center gap-1.5">
                <CalendarIcon className="size-4" />
                <span>Opens: {formatPickWindowDate(tournament.pick_window_opens_at)}</span>
              </div>
              <div className="flex items-center gap-1.5">
                <ClockIcon className="size-4" />
                <span>Closes: {formatPickWindowDate(tournament.pick_window_closes_at)}</span>
              </div>
            </div>
          )}

          {!isWindowOpen ? (
            <div className="flex flex-col items-center justify-center rounded-lg bg-muted/50 py-12">
              <LockIcon className="mb-3 size-10 text-muted-foreground" />
              <p className="text-center text-muted-foreground">{getClosedReason()}</p>
            </div>
          ) : (
            <GolferSelector
              golfers={golfersData?.golfers ?? []}
              onSelect={handleSelectGolfer}
              isLoading={golfersLoading}
            />
          )}
        </CardContent>
      </Card>

      <PickConfirmDialog
        golfer={selectedGolfer}
        tournamentName={tournament.name}
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        onConfirm={handleConfirmPick}
        isSubmitting={createPick.isPending || updatePick.isPending}
        isChanging={isChanging}
      />
    </>
  );
}
