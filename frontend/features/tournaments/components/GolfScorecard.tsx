"use client";

import {
  calculateBackNine,
  calculateFrontNine,
  getHoleScoreClass,
  getRoundLabel,
} from "./leaderboard-utils";
import { Button } from "@/components/shadcn/button";
import type { CourseInfo, HoleScore, RoundScore } from "@/features/tournaments/types";
import { cn } from "@/lib/utils";
import { ChevronUpIcon } from "lucide-react";

type GolfScorecardProps = {
  rounds: RoundScore[];
  onClose: () => void;
};

interface RoundGroup {
  course: CourseInfo | null;
  rounds: RoundScore[];
}

function groupRoundsByCourse(rounds: RoundScore[]): RoundGroup[] {
  const sortedRounds = [...rounds].sort((a, b) => a.round_number - b.round_number);
  const groups: RoundGroup[] = [];

  for (const round of sortedRounds) {
    const currentCourse = round.course ?? null;
    const lastGroup = groups.at(-1);

    if (lastGroup !== undefined && lastGroup.course?.id === currentCourse?.id) {
      lastGroup.rounds.push(round);
    } else {
      groups.push({ course: currentCourse, rounds: [round] });
    }
  }

  return groups;
}

function hasMultipleCourses(rounds: RoundScore[]): boolean {
  const courseIds = new Set<string | undefined>();
  for (const round of rounds) {
    courseIds.add(round.course?.id);
  }
  return courseIds.size > 1;
}

function getParForHoleInGroup(rounds: RoundScore[], holeNumber: number): number {
  for (const round of rounds) {
    const hole = round.holes?.find((h) => h.hole_number === holeNumber);
    if (hole) return hole.par;
  }
  return 4;
}

function getTotalParForGroup(rounds: RoundScore[]): number {
  const roundWithHoles = rounds.find((r) => r.holes && r.holes.length === 18);
  if (!roundWithHoles?.holes) return 72;
  return roundWithHoles.holes.reduce((sum, h) => sum + h.par, 0);
}

function getFrontNineParForGroup(rounds: RoundScore[]): number {
  const roundWithHoles = rounds.find((r) => r.holes && r.holes.length >= 9);
  if (!roundWithHoles?.holes) return 36;
  return roundWithHoles.holes.filter((h) => h.hole_number <= 9).reduce((sum, h) => sum + h.par, 0);
}

function getBackNineParForGroup(rounds: RoundScore[]): number {
  const roundWithHoles = rounds.find((r) => r.holes && r.holes.length === 18);
  if (!roundWithHoles?.holes) return 36;
  return roundWithHoles.holes.filter((h) => h.hole_number > 9).reduce((sum, h) => sum + h.par, 0);
}

export function GolfScorecard({ rounds, onClose }: GolfScorecardProps) {
  const holes = Array.from({ length: 18 }, (_, i) => i + 1);
  const frontNine = holes.slice(0, 9);
  const backNine = holes.slice(9);

  const groups = groupRoundsByCourse(rounds);
  const showCourseHeaders = hasMultipleCourses(rounds);

  const cellClass = "px-2 py-1 text-center font-mono text-xs";
  const headerCellClass = cn(cellClass, "bg-muted font-semibold");
  const columnCount = 1 + 9 + 1 + 9 + 1 + 1; // Hole + front9 + OUT + back9 + IN + TOT

  return (
    <div className="py-2">
      <div className="overflow-x-auto rounded border">
        <table className="w-full border-collapse text-sm">
          <thead>
            <tr className="border-b">
              <th className={cn(headerCellClass, "sticky left-0 z-10 bg-muted text-left")}>Hole</th>
              {frontNine.map((hole) => (
                <th key={hole} className={headerCellClass}>
                  {hole}
                </th>
              ))}
              <th className={cn(headerCellClass, "bg-muted/80")}>OUT</th>
              {backNine.map((hole) => (
                <th key={hole} className={headerCellClass}>
                  {hole}
                </th>
              ))}
              <th className={cn(headerCellClass, "bg-muted/80")}>IN</th>
              <th className={cn(headerCellClass, "bg-muted/80")}>TOT</th>
            </tr>
          </thead>
          <tbody>
            {groups.map((group, groupIndex) => (
              <CourseGroup
                key={group.course?.id ?? `group-${groupIndex}`}
                group={group}
                holes={holes}
                frontNine={frontNine}
                backNine={backNine}
                showCourseHeader={showCourseHeaders}
                columnCount={columnCount}
                cellClass={cellClass}
              />
            ))}
          </tbody>
        </table>
      </div>
      <div className="mt-2 flex justify-center">
        <Button variant="ghost" size="sm" onClick={onClose} className="w-full gap-1">
          <ChevronUpIcon className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}

interface CourseGroupProps {
  group: RoundGroup;
  holes: number[];
  frontNine: number[];
  backNine: number[];
  showCourseHeader: boolean;
  columnCount: number;
  cellClass: string;
}

function CourseGroup({
  group,
  holes,
  frontNine,
  backNine,
  showCourseHeader,
  columnCount,
  cellClass,
}: CourseGroupProps) {
  const courseName = group.course?.name ?? "Unknown Course";
  const coursePar = group.course?.par;

  return (
    <>
      {showCourseHeader && (
        <tr className="border-b bg-muted/40">
          <td
            colSpan={columnCount}
            className="px-2 py-1.5 text-left text-xs font-medium text-muted-foreground"
          >
            <span>{courseName}</span>
            {coursePar && <span className="ml-2 text-muted-foreground/70">Par {coursePar}</span>}
          </td>
        </tr>
      )}
      <tr className="border-b bg-muted/30">
        <td className={cn(cellClass, "sticky left-0 z-10 bg-muted/30 text-left font-medium")}>
          Par
        </td>
        {frontNine.map((hole) => (
          <td key={hole} className={cellClass}>
            {getParForHoleInGroup(group.rounds, hole)}
          </td>
        ))}
        <td className={cn(cellClass, "bg-muted/50 font-medium")}>
          {getFrontNineParForGroup(group.rounds)}
        </td>
        {backNine.map((hole) => (
          <td key={hole} className={cellClass}>
            {getParForHoleInGroup(group.rounds, hole)}
          </td>
        ))}
        <td className={cn(cellClass, "bg-muted/50 font-medium")}>
          {getBackNineParForGroup(group.rounds)}
        </td>
        <td className={cn(cellClass, "bg-muted/50 font-medium")}>
          {getTotalParForGroup(group.rounds)}
        </td>
      </tr>
      {group.rounds.map((round) => (
        <RoundRow key={round.round_number} round={round} holes={holes} />
      ))}
    </>
  );
}

interface RoundRowProps {
  round: RoundScore;
  holes: number[];
}

function RoundRow({ round, holes }: RoundRowProps) {
  const frontNine = holes.slice(0, 9);
  const backNine = holes.slice(9);

  const getHoleScore = (holeNumber: number): HoleScore | undefined => {
    return round.holes?.find((h) => h.hole_number === holeNumber);
  };

  const frontNineCalc = round.holes ? calculateFrontNine(round.holes) : null;
  const backNineCalc = round.holes ? calculateBackNine(round.holes) : null;

  const totalStrokes =
    frontNineCalc && backNineCalc ? frontNineCalc.strokes + backNineCalc.strokes : null;

  const cellClass = "px-2 py-1 text-center font-mono text-xs";

  return (
    <tr className="border-b last:border-b-0">
      <td className={cn(cellClass, "sticky left-0 z-10 bg-background text-left font-medium")}>
        {getRoundLabel(round.round_number)}
      </td>
      {frontNine.map((holeNum) => {
        const hole = getHoleScore(holeNum);
        return (
          <td
            key={holeNum}
            className={cn(cellClass, hole && getHoleScoreClass(hole.score, hole.par))}
          >
            {hole?.score ?? "-"}
          </td>
        );
      })}
      <td className={cn(cellClass, "bg-muted/30 font-medium")}>{frontNineCalc?.strokes ?? "-"}</td>
      {backNine.map((holeNum) => {
        const hole = getHoleScore(holeNum);
        return (
          <td
            key={holeNum}
            className={cn(cellClass, hole && getHoleScoreClass(hole.score, hole.par))}
          >
            {hole?.score ?? "-"}
          </td>
        );
      })}
      <td className={cn(cellClass, "bg-muted/30 font-medium")}>{backNineCalc?.strokes ?? "-"}</td>
      <td className={cn(cellClass, "bg-muted/30 font-medium")}>{totalStrokes ?? "-"}</td>
    </tr>
  );
}
