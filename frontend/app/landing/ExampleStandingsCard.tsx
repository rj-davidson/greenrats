import { ExampleCard } from "./ExampleCard";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/shadcn/table";
import { cn } from "@/lib/utils";
import { TrophyIcon, UsersIcon } from "lucide-react";

const EXAMPLE_STANDINGS = [
  { rank: 1, name: "JohnnyGolf", earnings: 12_450_000, isYou: false },
  { rank: 2, name: "BirdieQueen", earnings: 11_200_000, isYou: false },
  { rank: 3, name: "FairwayFred", earnings: 9_900_000, isYou: true },
  { rank: 4, name: "EagleEye", earnings: 8_750_000, isYou: false },
  { rank: 5, name: "PuttMaster", earnings: 7_300_000, isYou: false },
];

function formatEarnings(amount: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(amount);
}

function getRankDisplay(rank: number) {
  if (rank === 1) {
    return (
      <span className="flex items-center gap-1 text-amber-500">
        <TrophyIcon className="size-4" />1
      </span>
    );
  }
  if (rank === 2) {
    return <span className="text-slate-400">{rank}</span>;
  }
  if (rank === 3) {
    return <span className="text-amber-700">{rank}</span>;
  }
  return rank;
}

export function ExampleStandingsCard() {
  return (
    <ExampleCard title="Standings" icon={<UsersIcon className="size-4" />}>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-12">Rank</TableHead>
            <TableHead>Player</TableHead>
            <TableHead className="text-right">Earnings</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {EXAMPLE_STANDINGS.map((entry) => (
            <TableRow key={entry.name} className={cn(entry.isYou && "bg-primary/5")}>
              <TableCell className="font-medium">{getRankDisplay(entry.rank)}</TableCell>
              <TableCell>
                {entry.name}
                {entry.isYou && <span className="ml-1 text-xs text-muted-foreground">(you)</span>}
              </TableCell>
              <TableCell className="text-right">{formatEarnings(entry.earnings)}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </ExampleCard>
  );
}
