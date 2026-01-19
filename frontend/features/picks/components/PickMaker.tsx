"use client";

import { useState } from "react";
import { useAvailableGolfers, useCreatePick, usePickWindow } from "../queries";
import type { AvailableGolfer } from "../types";
import { GolferSelector } from "./GolferSelector";
import { PickConfirmDialog } from "./PickConfirmDialog";
import { Badge } from "@/components/shadcn/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/shadcn/card";
import { Skeleton } from "@/components/shadcn/skeleton";
import { CalendarIcon, ClockIcon, LockIcon } from "lucide-react";
import { toast } from "sonner";

interface PickMakerProps {
  leagueId: string;
  tournamentId: string;
  onPickSuccess?: () => void;
}

function formatDate(dateString: string) {
  return new Date(dateString).toLocaleDateString("en-US", {
    weekday: "short",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

export function PickMaker({ leagueId, tournamentId, onPickSuccess }: PickMakerProps) {
  const [selectedGolfer, setSelectedGolfer] = useState<AvailableGolfer | null>(null);
  const [confirmOpen, setConfirmOpen] = useState(false);

  const { data: pickWindow, isLoading: windowLoading } = usePickWindow(tournamentId);
  const { data: golfersData, isLoading: golfersLoading } = useAvailableGolfers(
    leagueId,
    tournamentId
  );
  const createPick = useCreatePick();

  const handleSelectGolfer = (golfer: AvailableGolfer) => {
    setSelectedGolfer(golfer);
    setConfirmOpen(true);
  };

  const handleConfirmPick = async () => {
    if (!selectedGolfer) return;

    try {
      await createPick.mutateAsync({
        tournament_id: tournamentId,
        golfer_id: selectedGolfer.id,
        league_id: leagueId,
      });
      toast.success(`Successfully picked ${selectedGolfer.name}!`);
      setConfirmOpen(false);
      setSelectedGolfer(null);
      onPickSuccess?.();
    } catch {
      toast.error("Failed to submit pick. Please try again.");
    }
  };

  if (windowLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-4 w-64" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-64 w-full" />
        </CardContent>
      </Card>
    );
  }

  if (!pickWindow) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground">Unable to load pick window information.</p>
        </CardContent>
      </Card>
    );
  }

  const isWindowOpen = pickWindow.is_open;

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-start justify-between gap-2">
            <div>
              <CardTitle>{pickWindow.tournament_name}</CardTitle>
              <CardDescription>Select your golfer for this tournament</CardDescription>
            </div>
            <Badge variant={isWindowOpen ? "default" : "secondary"}>
              {isWindowOpen ? "Open" : "Closed"}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="text-muted-foreground flex flex-wrap gap-4 text-sm">
            <div className="flex items-center gap-1.5">
              <CalendarIcon className="size-4" />
              <span>Opens: {formatDate(pickWindow.opens_at)}</span>
            </div>
            <div className="flex items-center gap-1.5">
              <ClockIcon className="size-4" />
              <span>Closes: {formatDate(pickWindow.closes_at)}</span>
            </div>
          </div>

          {!isWindowOpen ? (
            <div className="bg-muted/50 flex flex-col items-center justify-center rounded-lg py-12">
              <LockIcon className="text-muted-foreground mb-3 size-10" />
              <p className="text-muted-foreground text-center">
                {pickWindow.reason || "Pick window is closed"}
              </p>
            </div>
          ) : (
            <GolferSelector
              golfers={golfersData?.golfers ?? []}
              selectedGolferId={selectedGolfer?.id}
              onSelect={handleSelectGolfer}
              isLoading={golfersLoading}
            />
          )}
        </CardContent>
      </Card>

      <PickConfirmDialog
        golfer={selectedGolfer}
        tournamentName={pickWindow.tournament_name}
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        onConfirm={handleConfirmPick}
        isSubmitting={createPick.isPending}
      />
    </>
  );
}
