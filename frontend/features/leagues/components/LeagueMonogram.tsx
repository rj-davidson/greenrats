import type { League } from "../types";

const MONOGRAM_COLORS = [
  "bg-red-700",
  "bg-orange-700",
  "bg-amber-700",
  "bg-yellow-700",
  "bg-lime-700",
  "bg-green-700",
  "bg-emerald-700",
  "bg-teal-700",
  "bg-cyan-700",
  "bg-sky-700",
  "bg-blue-700",
  "bg-violet-700",
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
      className={`flex items-center justify-center text-white ${bgColor} ${className}`}
      style={{
        width: size,
        height: size,
        fontSize,
        borderRadius,
      }}
    >
      <span className="font-serif font-semibold leading-none">{initials}</span>
    </div>
  );
}
