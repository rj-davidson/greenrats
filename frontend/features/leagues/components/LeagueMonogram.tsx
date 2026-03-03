import type { League } from "@/features/leagues/types";

const MONOGRAM_COLORS = [
  "bg-chart-1",
  "bg-chart-2",
  "bg-chart-3",
  "bg-chart-4",
  "bg-chart-5",
] as const;

function getColorFromUUID(uuid: string): string {
  let hash = 0;
  for (let i = 0; i < uuid.length; i++) {
    hash = (hash << 5) - hash + uuid.charCodeAt(i);
    hash |= 0;
  }
  const index = Math.abs(hash) % MONOGRAM_COLORS.length;
  return MONOGRAM_COLORS[index];
}

function getInitials(name: string): string {
  const words = name.trim().split(/\s+/);
  if (words.length === 1) {
    return words[0].substring(0, 2).toUpperCase();
  }
  return (words[0][0] + words[1][0]).toUpperCase();
}

interface LeagueMonogramProps {
  league: Pick<League, "id" | "name">;
  size?: number;
  className?: string;
}

export function LeagueMonogram({ league, size = 32, className = "" }: LeagueMonogramProps) {
  const bgColor = getColorFromUUID(league.id);
  const initials = getInitials(league.name);
  const fontSize = Math.round(size * 0.45);
  const borderRadius = Math.round(size * 0.15);

  return (
    <div
      className={`flex shrink-0 items-center justify-center text-white ${bgColor} ${className}`}
      style={{
        width: size,
        height: size,
        fontSize,
        borderRadius,
      }}
    >
      <span className="font-serif leading-none font-semibold">{initials}</span>
    </div>
  );
}
