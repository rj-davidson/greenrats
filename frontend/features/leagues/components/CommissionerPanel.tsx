"use client";

import { useRegenerateJoinCode, useSetJoiningEnabled } from "@/features/leagues/queries";
import type { League } from "@/features/leagues/types";
import { PickManagement } from "@/features/leagues/components/PickManagement";
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
import { CheckIcon, CopyIcon, RefreshCwIcon, ShieldIcon, UsersIcon } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

interface CommissionerPanelProps {
  league: League;
}

export function CommissionerPanel({ league }: CommissionerPanelProps) {
  const [copied, setCopied] = useState(false);
  const [confirmRegenerateOpen, setConfirmRegenerateOpen] = useState(false);

  const regenerateCode = useRegenerateJoinCode();
  const setJoiningEnabled = useSetJoiningEnabled();

  const handleCopyCode = async () => {
    await navigator.clipboard.writeText(league.code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleRegenerateCode = async () => {
    try {
      await regenerateCode.mutateAsync(league.id);
      toast.success("Join code regenerated successfully");
      setConfirmRegenerateOpen(false);
    } catch {
      toast.error("Failed to regenerate join code");
    }
  };

  const handleToggleJoining = async () => {
    try {
      await setJoiningEnabled.mutateAsync({
        leagueId: league.id,
        enabled: !league.joining_enabled,
      });
      toast.success(league.joining_enabled ? "Joining disabled" : "Joining enabled");
    } catch {
      toast.error("Failed to update joining status");
    }
  };

  if (league.role !== "owner") {
    return null;
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center gap-2">
            <ShieldIcon className="size-5 text-primary" />
            <CardTitle className="text-base">Commissioner Controls</CardTitle>
          </div>
          <CardDescription>Manage league settings and membership</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Join Code</label>
            <div className="flex items-center gap-2">
              <code className="flex-1 rounded bg-muted px-3 py-2 font-mono text-lg tracking-wider">
                {league.code}
              </code>
              <Button variant="outline" size="icon" onClick={handleCopyCode}>
                {copied ? <CheckIcon className="size-4" /> : <CopyIcon className="size-4" />}
              </Button>
              <Button
                variant="outline"
                size="icon"
                onClick={() => setConfirmRegenerateOpen(true)}
                disabled={regenerateCode.isPending}
              >
                <RefreshCwIcon className="size-4" />
              </Button>
            </div>
          </div>

          <div className="flex items-center justify-between rounded-lg border p-3">
            <div className="flex items-center gap-2">
              <UsersIcon className="size-4 text-muted-foreground" />
              <span className="text-sm">Allow new members to join</span>
            </div>
            <Button
              variant={league.joining_enabled ? "default" : "outline"}
              size="sm"
              onClick={handleToggleJoining}
              disabled={setJoiningEnabled.isPending}
            >
              {league.joining_enabled ? "Enabled" : "Disabled"}
            </Button>
          </div>

          {league.member_count !== undefined && (
            <div className="text-sm text-muted-foreground">
              {league.member_count} member{league.member_count !== 1 ? "s" : ""} in this league
            </div>
          )}
        </CardContent>
      </Card>

      <PickManagement leagueId={league.id} />

      <Dialog open={confirmRegenerateOpen} onOpenChange={setConfirmRegenerateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Regenerate Join Code?</DialogTitle>
            <DialogDescription>
              This will create a new join code. The old code will no longer work. Anyone who has the
              old code will need the new one to join.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setConfirmRegenerateOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleRegenerateCode} disabled={regenerateCode.isPending}>
              {regenerateCode.isPending ? "Regenerating..." : "Regenerate"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
