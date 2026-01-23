import type { HoleScore, RoundScore } from "@/features/tournaments/types";

export function formatScoreToPar(score: number): string {
  if (score === 0) return "E";
  return score > 0 ? `+${score}` : `${score}`;
}

export function formatThru(thru: number, status: string): string {
  if (status === "finished") return "F";
  if (thru === 0) return "-";
  if (thru === 18) return "F";
  return `${thru}`;
}

export function getRoundLabel(roundNumber: number): string {
  return `R${roundNumber}`;
}

export function getCurrentRoundScore(entry: {
  rounds: RoundScore[];
  current_round: number;
}): string {
  const round = entry.rounds.find((r) => r.round_number === entry.current_round);
  if (!round || round.par_relative_score === null) return "-";
  return formatScoreToPar(round.par_relative_score);
}

export function calculateFrontNine(holes: HoleScore[]): { strokes: number; par: number } | null {
  const front = holes.filter((h) => h.hole_number <= 9);
  if (front.length === 0) return null;

  const played = front.filter((h) => h.score !== null);
  if (played.length === 0) return null;

  return {
    strokes: played.reduce((sum, h) => sum + (h.score ?? 0), 0),
    par: played.reduce((sum, h) => sum + h.par, 0),
  };
}

export function calculateBackNine(holes: HoleScore[]): { strokes: number; par: number } | null {
  const back = holes.filter((h) => h.hole_number > 9);
  if (back.length === 0) return null;

  const played = back.filter((h) => h.score !== null);
  if (played.length === 0) return null;

  return {
    strokes: played.reduce((sum, h) => sum + (h.score ?? 0), 0),
    par: played.reduce((sum, h) => sum + h.par, 0),
  };
}

export function getHoleScoreClass(score: number | null, par: number): string {
  if (score === null) return "";
  const diff = score - par;
  if (diff <= -2) return "text-amber-600 dark:text-amber-400 font-semibold";
  if (diff === -1) return "text-primary";
  if (diff === 1) return "text-destructive";
  if (diff >= 2) return "text-destructive font-semibold";
  return "";
}
