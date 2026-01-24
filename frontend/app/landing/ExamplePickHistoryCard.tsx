import { ExampleCard } from "./ExampleCard";
import { HistoryIcon } from "lucide-react";

const EXAMPLE_PICKS = [
  { tournament: "The Masters", golfer: "Scottie Scheffler", earnings: 3_600_000 },
  { tournament: "PGA Championship", golfer: "Xander Schauffele", earnings: 3_300_000 },
  { tournament: "The Open Championship", golfer: "Brian Harman", earnings: 3_000_000 },
];

function formatEarnings(amount: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(amount);
}

export function ExamplePickHistoryCard() {
  const totalEarnings = EXAMPLE_PICKS.reduce((sum, p) => sum + p.earnings, 0);

  return (
    <ExampleCard title="Pick History" icon={<HistoryIcon className="size-4" />}>
      <div className="space-y-3">
        {EXAMPLE_PICKS.map((pick) => (
          <div key={pick.tournament} className="space-y-0.5">
            <div className="text-sm font-medium">{pick.tournament}</div>
            <div className="flex items-center justify-between text-sm text-muted-foreground">
              <span>{pick.golfer}</span>
              <span className="font-medium text-foreground">{formatEarnings(pick.earnings)}</span>
            </div>
          </div>
        ))}
        <div className="flex items-center justify-between border-t pt-3">
          <span className="font-semibold">Total</span>
          <span className="font-semibold">{formatEarnings(totalEarnings)}</span>
        </div>
      </div>
    </ExampleCard>
  );
}
