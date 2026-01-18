import { TournamentSchedule } from "@/features/tournaments/components";

export default function TournamentsPage() {
  return (
    <div className="container mx-auto p-8">
      <div className="mb-8">
        <h1 className="mb-2 text-3xl font-bold">Tournaments</h1>
        <p className="text-muted-foreground">PGA Tour schedule and tournament details</p>
      </div>
      <TournamentSchedule />
    </div>
  );
}
