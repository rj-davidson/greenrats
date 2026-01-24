import { ExampleCard } from "./ExampleCard";
import { Progress } from "@/components/shadcn/progress";
import { CalendarDaysIcon } from "lucide-react";

export function ExampleSeasonProgressCard() {
  const completedCount = 18;
  const total = 42;
  const progress = (completedCount / total) * 100;
  const currentTournament = "Truist Championship";

  return (
    <ExampleCard title="Season Progress" icon={<CalendarDaysIcon className="size-4" />}>
      <div className="space-y-4">
        <div>
          <div className="mb-2 flex items-baseline justify-between">
            <span className="text-2xl font-bold">{completedCount}</span>
            <span className="text-sm text-muted-foreground">of {total} tournaments</span>
          </div>
          <Progress value={progress} />
        </div>
        <div className="rounded-lg bg-muted/50 p-3">
          <p className="text-xs text-muted-foreground">Current</p>
          <p className="font-medium">{currentTournament}</p>
        </div>
      </div>
    </ExampleCard>
  );
}
