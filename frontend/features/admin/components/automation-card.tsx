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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/shadcn/select";
import {
  useAdminTournaments,
  useTriggerSyncTournaments,
  useTriggerSyncPlayers,
  useTriggerSyncLeaderboard,
  useTriggerSyncEarnings,
  useTriggerSyncField,
} from "@/features/admin/queries";
import { RefreshCwIcon } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

export function AutomationsPage() {
  const { data: tournamentsData, isLoading: loadingTournaments } = useAdminTournaments();
  const syncTournaments = useTriggerSyncTournaments();
  const syncPlayers = useTriggerSyncPlayers();
  const syncLeaderboard = useTriggerSyncLeaderboard();
  const syncEarnings = useTriggerSyncEarnings();
  const syncField = useTriggerSyncField();

  const [selectedTournament, setSelectedTournament] = useState<string>("");

  const handleSyncTournaments = async () => {
    try {
      await syncTournaments.mutateAsync();
      toast.success("Tournament sync started");
    } catch {
      toast.error("Failed to start tournament sync");
    }
  };

  const handleSyncPlayers = async () => {
    try {
      await syncPlayers.mutateAsync();
      toast.success("Player sync started");
    } catch {
      toast.error("Failed to start player sync");
    }
  };

  const handleSyncLeaderboard = async () => {
    if (!selectedTournament) {
      toast.error("Please select a tournament");
      return;
    }
    try {
      await syncLeaderboard.mutateAsync(selectedTournament);
      toast.success("Leaderboard sync started");
    } catch {
      toast.error("Failed to start leaderboard sync");
    }
  };

  const handleSyncEarnings = async () => {
    if (!selectedTournament) {
      toast.error("Please select a tournament");
      return;
    }
    try {
      await syncEarnings.mutateAsync(selectedTournament);
      toast.success("Earnings sync started");
    } catch {
      toast.error("Failed to start earnings sync");
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

  return (
    <div className="space-y-6">
      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Sync Tournaments</CardTitle>
            <CardDescription>
              Fetch and update all tournament data from BallDontLie API
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button
              onClick={handleSyncTournaments}
              disabled={syncTournaments.isPending}
              className="w-full"
            >
              <RefreshCwIcon className={syncTournaments.isPending ? "animate-spin" : ""} />
              {syncTournaments.isPending ? "Syncing..." : "Sync Tournaments"}
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Sync Players</CardTitle>
            <CardDescription>Fetch and update all golfer data from BallDontLie API</CardDescription>
          </CardHeader>
          <CardContent>
            <Button onClick={handleSyncPlayers} disabled={syncPlayers.isPending} className="w-full">
              <RefreshCwIcon className={syncPlayers.isPending ? "animate-spin" : ""} />
              {syncPlayers.isPending ? "Syncing..." : "Sync Players"}
            </Button>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Tournament-Specific Actions</CardTitle>
          <CardDescription>
            Sync leaderboard or earnings data for a specific tournament
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <Select value={selectedTournament} onValueChange={setSelectedTournament}>
            <SelectTrigger>
              <SelectValue
                placeholder={loadingTournaments ? "Loading..." : "Select a tournament"}
              />
            </SelectTrigger>
            <SelectContent>
              {tournamentsData?.tournaments.map((tournament) => (
                <SelectItem key={tournament.id} value={tournament.id}>
                  {tournament.name} ({tournament.status})
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <div className="grid gap-4 md:grid-cols-3">
            <Button
              onClick={handleSyncField}
              disabled={!selectedTournament || syncField.isPending}
              variant="outline"
              className="w-full"
            >
              <RefreshCwIcon className={syncField.isPending ? "animate-spin" : ""} />
              {syncField.isPending ? "Syncing..." : "Sync Field"}
            </Button>

            <Button
              onClick={handleSyncLeaderboard}
              disabled={!selectedTournament || syncLeaderboard.isPending}
              variant="outline"
              className="w-full"
            >
              <RefreshCwIcon className={syncLeaderboard.isPending ? "animate-spin" : ""} />
              {syncLeaderboard.isPending ? "Syncing..." : "Sync Leaderboard"}
            </Button>

            <Button
              onClick={handleSyncEarnings}
              disabled={!selectedTournament || syncEarnings.isPending}
              variant="outline"
              className="w-full"
            >
              <RefreshCwIcon className={syncEarnings.isPending ? "animate-spin" : ""} />
              {syncEarnings.isPending ? "Syncing..." : "Sync Earnings"}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
