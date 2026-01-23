"use client";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/shadcn/avatar";
import { Button } from "@/components/shadcn/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/shadcn/dialog";
import type { AvailableGolfer } from "@/features/picks/types";
import { TriangleAlertIcon } from "lucide-react";

interface PickConfirmDialogProps {
  golfer: AvailableGolfer | null;
  tournamentName: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
  isSubmitting?: boolean;
  isChanging?: boolean;
}

export function PickConfirmDialog({
  golfer,
  tournamentName,
  open,
  onOpenChange,
  onConfirm,
  isSubmitting,
  isChanging,
}: PickConfirmDialogProps) {
  if (!golfer) return null;

  const initials = golfer.name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isChanging ? "Confirm Pick Change" : "Confirm Your Pick"}</DialogTitle>
          <DialogDescription>
            You are about to {isChanging ? "change your pick to" : "pick"}{" "}
            <strong>{golfer.name}</strong> for {tournamentName}.
          </DialogDescription>
        </DialogHeader>

        <div className="flex items-center gap-4 rounded-lg border p-4">
          <Avatar className="size-14">
            {golfer.image_url && <AvatarImage src={golfer.image_url} alt={golfer.name} />}
            <AvatarFallback>{initials}</AvatarFallback>
          </Avatar>
          <div>
            <div className="font-medium">{golfer.name}</div>
            <div className="text-sm text-muted-foreground">
              {golfer.country_code}
              {golfer.owgr && golfer.owgr > 0 && ` • OWGR #${golfer.owgr}`}
            </div>
          </div>
        </div>

        <div className="flex items-start gap-3 rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
          <TriangleAlertIcon className="mt-0.5 size-4 shrink-0" />
          <div>
            {isChanging ? (
              <>
                <strong>You can still change this pick</strong> until the pick window closes. Once
                confirmed, you cannot pick {golfer.name} again this season in this league.
              </>
            ) : (
              <>
                <strong>This cannot be changed.</strong> Once confirmed, you cannot pick{" "}
                {golfer.name} again this season in this league.
              </>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button onClick={onConfirm} disabled={isSubmitting}>
            {isSubmitting
              ? isChanging
                ? "Changing..."
                : "Submitting..."
              : isChanging
                ? "Change Pick"
                : "Confirm Pick"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
