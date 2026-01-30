"use client";

import {
  calculateBackNine,
  calculateFrontNine,
  getHoleScoreClass,
  getRoundLabel,
} from "./leaderboard-utils";
import { Button } from "@/components/shadcn/button";
import type { CourseInfo, HoleScore, RoundScore } from "@/features/tournaments/types";
import { useIsMobile } from "@/hooks/use-mobile";
import { cn } from "@/lib/utils";
import { ChevronUpIcon } from "lucide-react";

type GolfScorecardProps = {
  rounds: RoundScore[];
  onClose?: () => void;
};

interface RoundGroup {
  course: CourseInfo | null;
  abbreviation: string;
  rounds: RoundScore[];
}

interface CourseAbbreviation {
  course: CourseInfo;
  abbreviation: string;
}

function getParenContent(name: string): string | null {
  const match = name.match(/\(([^)]+)\)/);
  if (!match) return null;
  return match[1].replace(/\s+Course$/i, "").trim();
}

function generateCourseAbbreviations(courses: CourseInfo[]): Map<string, string> {
  const result = new Map<string, string>();
  if (courses.length === 0) return result;
  if (courses.length === 1) {
    result.set(courses[0].id, "");
    return result;
  }

  const parenContents = courses.map((c) => getParenContent(c.name));
  if (parenContents.every((p) => p !== null)) {
    const abbrevs = parenContents.map((p) => p[0].toUpperCase());
    if (new Set(abbrevs).size === abbrevs.length) {
      courses.forEach((c, i) => result.set(c.id, abbrevs[i]));
      return result;
    }
  }

  const words = courses.map((c) => c.name.split(/\s+/));
  const minLen = Math.min(...words.map((w) => w.length));

  for (let i = 0; i < minLen; i++) {
    const wordsAtPos = words.map((w) => w[i]);
    if (new Set(wordsAtPos).size === courses.length) {
      const abbrevs = wordsAtPos.map((w) => w[0].toUpperCase());
      if (new Set(abbrevs).size === abbrevs.length) {
        courses.forEach((c, idx) => result.set(c.id, abbrevs[idx]));
        return result;
      }
      const abbrevs2 = wordsAtPos.map((w) => w.slice(0, 2).toUpperCase());
      courses.forEach((c, idx) => result.set(c.id, abbrevs2[idx]));
      return result;
    }
  }

  courses.forEach((c, i) => result.set(c.id, String(i + 1)));
  return result;
}

function getUniqueCourses(rounds: RoundScore[]): CourseInfo[] {
  const seen = new Set<string>();
  const courses: CourseInfo[] = [];
  for (const round of rounds) {
    if (round.course && !seen.has(round.course.id)) {
      seen.add(round.course.id);
      courses.push(round.course);
    }
  }
  return courses;
}

function groupRoundsByCourse(
  rounds: RoundScore[],
  abbreviations: Map<string, string>,
): RoundGroup[] {
  const sortedRounds = [...rounds].sort((a, b) => a.round_number - b.round_number);
  const groups: RoundGroup[] = [];

  for (const round of sortedRounds) {
    const currentCourse = round.course ?? null;
    const abbrev = currentCourse ? (abbreviations.get(currentCourse.id) ?? "") : "";
    const lastGroup = groups.at(-1);

    if (lastGroup !== undefined && lastGroup.course?.id === currentCourse?.id) {
      lastGroup.rounds.push(round);
    } else {
      groups.push({ course: currentCourse, abbreviation: abbrev, rounds: [round] });
    }
  }

  return groups;
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
  const isMobile = useIsMobile();
  const courses = getUniqueCourses(rounds);
  const abbreviations = generateCourseAbbreviations(courses);
  const groups = groupRoundsByCourse(rounds, abbreviations);
  const isMultiCourse = courses.length > 1;

  const courseAbbreviations: CourseAbbreviation[] = courses.map((c) => ({
    course: c,
    abbreviation: abbreviations.get(c.id) ?? "",
  }));

  return (
    <div className={onClose ? "py-2" : ""}>
      {isMobile ? (
        <StackedScorecard
          groups={groups}
          isMultiCourse={isMultiCourse}
          courseAbbreviations={courseAbbreviations}
        />
      ) : (
        <WideScorecard
          groups={groups}
          isMultiCourse={isMultiCourse}
          courseAbbreviations={courseAbbreviations}
        />
      )}
      {onClose && (
        <div className="mt-2 flex justify-center">
          <Button variant="ghost" size="sm" onClick={onClose} className="w-full gap-1">
            <ChevronUpIcon className="h-4 w-4" />
          </Button>
        </div>
      )}
    </div>
  );
}

// ============================================================================
// Wide Layout (Desktop)
// ============================================================================

interface WideScorecardProps {
  groups: RoundGroup[];
  isMultiCourse: boolean;
  courseAbbreviations: CourseAbbreviation[];
}

function WideScorecard({ groups, isMultiCourse, courseAbbreviations }: WideScorecardProps) {
  const holes = Array.from({ length: 18 }, (_, i) => i + 1);
  const frontNine = holes.slice(0, 9);
  const backNine = holes.slice(9);

  const cellClass = "px-2 py-1 text-center font-mono text-xs";
  const headerCellClass = cn(cellClass, "bg-muted font-semibold");

  return (
    <div className="overflow-x-auto rounded border">
      <table className="w-full border-collapse text-sm">
        <thead>
          {isMultiCourse ? (
            <tr className="border-b bg-muted/40">
              <td colSpan={22} className="px-2 py-1.5 text-left text-xs text-muted-foreground">
                {courseAbbreviations.map((ca, i) => (
                  <span key={ca.course.id} className={i > 0 ? "ml-4" : ""}>
                    <span className="font-semibold">{ca.abbreviation}</span>
                    <span className="ml-1">= {ca.course.name}</span>
                    {ca.course.par && (
                      <span className="ml-1 text-muted-foreground/70">({ca.course.par})</span>
                    )}
                  </span>
                ))}
              </td>
            </tr>
          ) : groups[0]?.course ? (
            <tr className="border-b bg-muted/40">
              <td
                colSpan={22}
                className="px-2 py-1.5 text-left text-xs font-medium text-muted-foreground"
              >
                <span>{groups[0].course.name}</span>
                {groups[0].course.par && (
                  <span className="ml-2 text-muted-foreground/70">Par {groups[0].course.par}</span>
                )}
              </td>
            </tr>
          ) : null}
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
            <WideGroupRows
              key={group.course?.id ?? `group-${groupIndex}`}
              group={group}
              frontNine={frontNine}
              backNine={backNine}
              isMultiCourse={isMultiCourse}
              showSeparator={groupIndex > 0}
              cellClass={cellClass}
            />
          ))}
        </tbody>
      </table>
    </div>
  );
}

interface WideGroupRowsProps {
  group: RoundGroup;
  frontNine: number[];
  backNine: number[];
  isMultiCourse: boolean;
  showSeparator: boolean;
  cellClass: string;
}

function WideGroupRows({
  group,
  frontNine,
  backNine,
  isMultiCourse,
  showSeparator,
  cellClass,
}: WideGroupRowsProps) {
  const abbrevSuffix = isMultiCourse && group.abbreviation ? ` ${group.abbreviation}` : "";

  return (
    <>
      {showSeparator && (
        <tr className="border-t-2 border-muted-foreground/20">
          <td colSpan={22} className="h-1" />
        </tr>
      )}
      <tr className="border-b bg-muted/30">
        <td className={cn(cellClass, "sticky left-0 z-10 bg-muted/30 text-left font-medium")}>
          Par{abbrevSuffix}
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
        <WideRoundRow
          key={round.round_number}
          round={round}
          frontNine={frontNine}
          backNine={backNine}
          abbreviation={isMultiCourse ? group.abbreviation : ""}
          cellClass={cellClass}
        />
      ))}
    </>
  );
}

interface WideRoundRowProps {
  round: RoundScore;
  frontNine: number[];
  backNine: number[];
  abbreviation: string;
  cellClass: string;
}

function WideRoundRow({ round, frontNine, backNine, abbreviation, cellClass }: WideRoundRowProps) {
  const getHoleScore = (holeNumber: number): HoleScore | undefined => {
    return round.holes?.find((h) => h.hole_number === holeNumber);
  };

  const frontNineCalc = round.holes ? calculateFrontNine(round.holes) : null;
  const backNineCalc = round.holes ? calculateBackNine(round.holes) : null;
  const totalStrokes =
    frontNineCalc && backNineCalc ? frontNineCalc.strokes + backNineCalc.strokes : null;

  const label = abbreviation
    ? `${getRoundLabel(round.round_number)} ${abbreviation}`
    : getRoundLabel(round.round_number);

  return (
    <tr className="border-b last:border-b-0">
      <td className={cn(cellClass, "sticky left-0 z-10 bg-background text-left font-medium")}>
        {label}
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

// ============================================================================
// Stacked Layout (Mobile)
// ============================================================================

interface StackedScorecardProps {
  groups: RoundGroup[];
  isMultiCourse: boolean;
  courseAbbreviations: CourseAbbreviation[];
}

function StackedScorecard({ groups, isMultiCourse, courseAbbreviations }: StackedScorecardProps) {
  const frontNine = Array.from({ length: 9 }, (_, i) => i + 1);
  const backNine = Array.from({ length: 9 }, (_, i) => i + 10);

  const cellClass = "px-1.5 py-1 text-center font-mono text-xs";
  const headerCellClass = cn(cellClass, "bg-muted font-semibold");
  const columnCount = 1 + 9 + 1 + 1; // Label + 9 holes + subtotal + total

  return (
    <div className="overflow-x-auto rounded border">
      <table className="w-full border-collapse text-sm">
        <thead>
          {isMultiCourse ? (
            <tr className="border-b bg-muted/40">
              <td
                colSpan={columnCount}
                className="px-2 py-1.5 text-left text-xs text-muted-foreground"
              >
                {courseAbbreviations.map((ca, i) => (
                  <span key={ca.course.id} className={i > 0 ? "ml-3" : ""}>
                    <span className="font-semibold">{ca.abbreviation}</span>
                    <span className="ml-1">= {ca.course.name.split("(")[0].trim()}</span>
                    {ca.course.par && (
                      <span className="ml-1 text-muted-foreground/70">({ca.course.par})</span>
                    )}
                  </span>
                ))}
              </td>
            </tr>
          ) : groups[0]?.course ? (
            <tr className="border-b bg-muted/40">
              <td
                colSpan={columnCount}
                className="px-2 py-1.5 text-left text-xs font-medium text-muted-foreground"
              >
                <span>{groups[0].course.name}</span>
                {groups[0].course.par && (
                  <span className="ml-2 text-muted-foreground/70">Par {groups[0].course.par}</span>
                )}
              </td>
            </tr>
          ) : null}
        </thead>
        <tbody>
          {groups.map((group, groupIndex) => (
            <StackedGroupRows
              key={group.course?.id ?? `group-${groupIndex}`}
              group={group}
              frontNine={frontNine}
              backNine={backNine}
              isMultiCourse={isMultiCourse}
              showSeparator={groupIndex > 0}
              cellClass={cellClass}
              headerCellClass={headerCellClass}
              columnCount={columnCount}
            />
          ))}
        </tbody>
      </table>
    </div>
  );
}

interface StackedGroupRowsProps {
  group: RoundGroup;
  frontNine: number[];
  backNine: number[];
  isMultiCourse: boolean;
  showSeparator: boolean;
  cellClass: string;
  headerCellClass: string;
  columnCount: number;
}

function StackedGroupRows({
  group,
  frontNine,
  backNine,
  isMultiCourse,
  showSeparator,
  cellClass,
  headerCellClass,
  columnCount,
}: StackedGroupRowsProps) {
  const abbrevSuffix = isMultiCourse && group.abbreviation ? ` ${group.abbreviation}` : "";

  const getHoleScore = (round: RoundScore, holeNumber: number): HoleScore | undefined => {
    return round.holes?.find((h) => h.hole_number === holeNumber);
  };

  return (
    <>
      {showSeparator && (
        <tr className="border-t-2 border-muted-foreground/20">
          <td colSpan={columnCount} className="h-1" />
        </tr>
      )}

      {/* ===== OUT SECTION ===== */}
      <tr className="border-b bg-muted/50">
        <td className={cn(headerCellClass, "sticky left-0 z-10 bg-muted/50 text-left")}>OUT</td>
        <td colSpan={9} className={headerCellClass}></td>
        <td className={cn(headerCellClass, "bg-muted/60")}></td>
        <td className={cn(headerCellClass, "bg-muted/60")}></td>
      </tr>
      <tr className="border-b">
        <td className={cn(headerCellClass, "sticky left-0 z-10 bg-muted text-left")}>Hole</td>
        {frontNine.map((hole) => (
          <td key={hole} className={headerCellClass}>
            {hole}
          </td>
        ))}
        <td className={cn(headerCellClass, "bg-muted/80")}></td>
        <td className={cn(headerCellClass, "bg-muted/80")}></td>
      </tr>
      <tr className="border-b bg-muted/30">
        <td className={cn(cellClass, "sticky left-0 z-10 bg-muted/30 text-left font-medium")}>
          Par{abbrevSuffix}
        </td>
        {frontNine.map((hole) => (
          <td key={hole} className={cellClass}>
            {getParForHoleInGroup(group.rounds, hole)}
          </td>
        ))}
        <td className={cn(cellClass, "bg-muted/50 font-medium")}>
          {getFrontNineParForGroup(group.rounds)}
        </td>
        <td className={cn(cellClass, "bg-muted/50")}></td>
      </tr>
      {group.rounds.map((round) => {
        const frontNineCalc = round.holes ? calculateFrontNine(round.holes) : null;
        const label = abbrevSuffix
          ? `${getRoundLabel(round.round_number)}${abbrevSuffix}`
          : getRoundLabel(round.round_number);
        return (
          <tr key={`front-${round.round_number}`} className="border-b">
            <td className={cn(cellClass, "sticky left-0 z-10 bg-background text-left font-medium")}>
              {label}
            </td>
            {frontNine.map((holeNum) => {
              const hole = getHoleScore(round, holeNum);
              return (
                <td
                  key={holeNum}
                  className={cn(cellClass, hole && getHoleScoreClass(hole.score, hole.par))}
                >
                  {hole?.score ?? "-"}
                </td>
              );
            })}
            <td className={cn(cellClass, "bg-muted/30 font-medium")}>
              {frontNineCalc?.strokes ?? "-"}
            </td>
            <td className={cn(cellClass, "bg-muted/30")}></td>
          </tr>
        );
      })}

      {/* ===== IN SECTION ===== */}
      <tr className="border-t-2 border-b border-muted-foreground/10 bg-muted/50">
        <td className={cn(headerCellClass, "sticky left-0 z-10 bg-muted/50 text-left")}>IN</td>
        <td colSpan={9} className={headerCellClass}></td>
        <td className={cn(headerCellClass, "bg-muted/60")}></td>
        <td className={cn(headerCellClass, "bg-muted/60")}>TOT</td>
      </tr>
      <tr className="border-b">
        <td className={cn(headerCellClass, "sticky left-0 z-10 bg-muted text-left")}>Hole</td>
        {backNine.map((hole) => (
          <td key={hole} className={headerCellClass}>
            {hole}
          </td>
        ))}
        <td className={cn(headerCellClass, "bg-muted/80")}></td>
        <td className={cn(headerCellClass, "bg-muted/80")}></td>
      </tr>
      <tr className="border-b bg-muted/30">
        <td className={cn(cellClass, "sticky left-0 z-10 bg-muted/30 text-left font-medium")}>
          Par{abbrevSuffix}
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
      {group.rounds.map((round) => {
        const frontNineCalc = round.holes ? calculateFrontNine(round.holes) : null;
        const backNineCalc = round.holes ? calculateBackNine(round.holes) : null;
        const totalStrokes =
          frontNineCalc && backNineCalc ? frontNineCalc.strokes + backNineCalc.strokes : null;
        const label = abbrevSuffix
          ? `${getRoundLabel(round.round_number)}${abbrevSuffix}`
          : getRoundLabel(round.round_number);
        return (
          <tr key={`back-${round.round_number}`} className="border-b last:border-b-0">
            <td className={cn(cellClass, "sticky left-0 z-10 bg-background text-left font-medium")}>
              {label}
            </td>
            {backNine.map((holeNum) => {
              const hole = getHoleScore(round, holeNum);
              return (
                <td
                  key={holeNum}
                  className={cn(cellClass, hole && getHoleScoreClass(hole.score, hole.par))}
                >
                  {hole?.score ?? "-"}
                </td>
              );
            })}
            <td className={cn(cellClass, "bg-muted/30 font-medium")}>
              {backNineCalc?.strokes ?? "-"}
            </td>
            <td className={cn(cellClass, "bg-muted/30 font-medium")}>{totalStrokes ?? "-"}</td>
          </tr>
        );
      })}
    </>
  );
}
