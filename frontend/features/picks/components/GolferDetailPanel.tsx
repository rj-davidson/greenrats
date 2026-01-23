"use client";

import type { GolferBio, GolferSeasonStats } from "@/features/picks/types";
import { CalendarIcon, GraduationCapIcon, MapPinIcon, TrophyIcon } from "lucide-react";

interface GolferDetailPanelProps {
  stats?: GolferSeasonStats | null;
  bio?: GolferBio | null;
}

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(amount);
}

function formatStat(value: number | null | undefined, decimals = 1, suffix = ""): string {
  if (value === null || value === undefined) return "-";
  return `${value.toFixed(decimals)}${suffix}`;
}

function formatBirthDate(dateStr: string | null | undefined): string | null {
  if (!dateStr) return null;
  const date = new Date(dateStr);
  const now = new Date();
  const age = now.getFullYear() - date.getFullYear();
  const month = date.toLocaleDateString("en-US", { month: "short" });
  const day = date.getDate();
  const year = date.getFullYear();
  return `${month} ${day}, ${year} (${age})`;
}

function formatLocation(city?: string, state?: string, country?: string): string | null {
  const parts = [city, state, country].filter(Boolean);
  if (parts.length === 0) return null;
  return parts.join(", ");
}

export function GolferDetailPanel({ stats, bio }: GolferDetailPanelProps) {
  const hasAnyStats = stats != null && Object.values(stats).some((v) => v != null);
  const hasAnyBio = bio != null && Object.values(bio).some((v) => v != null && v !== "");

  if (!hasAnyStats && !hasAnyBio) {
    return (
      <div className="px-4 py-6 text-center text-sm text-muted-foreground">
        No additional information available for this golfer.
      </div>
    );
  }

  const birthplace = bio
    ? formatLocation(bio.birthplace_city, bio.birthplace_state, bio.birthplace_country)
    : null;
  const residence = bio
    ? formatLocation(bio.residence_city, bio.residence_state, bio.residence_country)
    : null;
  const birthDate = bio?.birth_date ? formatBirthDate(bio.birth_date) : null;

  return (
    <div className="grid gap-4 p-4 md:grid-cols-2">
      {stats != null && hasAnyStats && (
        <div className="space-y-3">
          <h4 className="flex items-center gap-2 text-sm font-medium">
            <TrophyIcon className="size-4" />
            Season Stats
          </h4>
          <div className="grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
            {stats.events_played != null && (
              <>
                <span className="text-muted-foreground">Events</span>
                <span className="text-right tabular-nums">{stats.events_played}</span>
              </>
            )}
            {stats.cuts_made != null && (
              <>
                <span className="text-muted-foreground">Cuts Made</span>
                <span className="text-right tabular-nums">{stats.cuts_made}</span>
              </>
            )}
            {stats.wins != null && (
              <>
                <span className="text-muted-foreground">Wins</span>
                <span className="text-right tabular-nums">{stats.wins}</span>
              </>
            )}
            {stats.top_10s != null && (
              <>
                <span className="text-muted-foreground">Top 10s</span>
                <span className="text-right tabular-nums">{stats.top_10s}</span>
              </>
            )}
            {stats.earnings != null && (
              <>
                <span className="text-muted-foreground">Earnings</span>
                <span className="text-right tabular-nums">{formatCurrency(stats.earnings)}</span>
              </>
            )}
            {stats.scoring_avg != null && (
              <>
                <span className="text-muted-foreground">Scoring Avg</span>
                <span className="text-right tabular-nums">{formatStat(stats.scoring_avg, 2)}</span>
              </>
            )}
            {stats.driving_distance != null && (
              <>
                <span className="text-muted-foreground">Driving Dist</span>
                <span className="text-right tabular-nums">
                  {formatStat(stats.driving_distance, 1)} yds
                </span>
              </>
            )}
            {stats.driving_accuracy != null && (
              <>
                <span className="text-muted-foreground">Driving Acc</span>
                <span className="text-right tabular-nums">
                  {formatStat(stats.driving_accuracy, 1)}%
                </span>
              </>
            )}
            {stats.gir_pct != null && (
              <>
                <span className="text-muted-foreground">GIR</span>
                <span className="text-right tabular-nums">{formatStat(stats.gir_pct, 1)}%</span>
              </>
            )}
            {stats.putting_avg != null && (
              <>
                <span className="text-muted-foreground">Putting Avg</span>
                <span className="text-right tabular-nums">{formatStat(stats.putting_avg, 2)}</span>
              </>
            )}
            {stats.scrambling_pct != null && (
              <>
                <span className="text-muted-foreground">Scrambling</span>
                <span className="text-right tabular-nums">
                  {formatStat(stats.scrambling_pct, 1)}%
                </span>
              </>
            )}
          </div>
        </div>
      )}

      {bio != null && hasAnyBio && (
        <div className="space-y-3">
          <h4 className="text-sm font-medium">Bio</h4>
          <div className="space-y-2 text-sm">
            {birthDate && (
              <div className="flex items-center gap-2">
                <CalendarIcon className="size-4 text-muted-foreground" />
                <span>{birthDate}</span>
              </div>
            )}
            {birthplace && (
              <div className="flex items-center gap-2">
                <MapPinIcon className="size-4 text-muted-foreground" />
                <span>From {birthplace}</span>
              </div>
            )}
            {residence && (
              <div className="flex items-center gap-2">
                <MapPinIcon className="size-4 text-muted-foreground" />
                <span>Lives in {residence}</span>
              </div>
            )}
            {bio.school && (
              <div className="flex items-center gap-2">
                <GraduationCapIcon className="size-4 text-muted-foreground" />
                <span>{bio.school}</span>
              </div>
            )}
            {bio.turned_pro && (
              <div className="flex items-center gap-2">
                <TrophyIcon className="size-4 text-muted-foreground" />
                <span>Turned Pro: {bio.turned_pro}</span>
              </div>
            )}
            {bio.height && <div className="text-muted-foreground">Height: {bio.height}</div>}
            {bio.weight && <div className="text-muted-foreground">Weight: {bio.weight}</div>}
          </div>
        </div>
      )}
    </div>
  );
}
