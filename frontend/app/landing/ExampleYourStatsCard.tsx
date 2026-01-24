import { ExampleCard } from "./ExampleCard";
import { TrophyIcon } from "lucide-react";

function formatEarnings(amount: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(amount);
}

export function ExampleYourStatsCard() {
  const rank = 3;
  const totalMembers = 12;
  const earnings = 9_900_000;
  const gapFromFirst = 2_550_000;

  return (
    <ExampleCard title="Your Stats" icon={<TrophyIcon className="size-4" />}>
      <div className="space-y-4">
        <div className="flex items-baseline gap-2">
          <span className="text-4xl font-bold">{rank}</span>
          <span className="text-sm text-muted-foreground">of {totalMembers} members</span>
        </div>
        <div>
          <p className="text-sm text-muted-foreground">Total Earnings</p>
          <p className="text-2xl font-semibold">{formatEarnings(earnings)}</p>
        </div>
        <p className="text-sm text-muted-foreground">
          {formatEarnings(gapFromFirst)} behind leader
        </p>
      </div>
    </ExampleCard>
  );
}
