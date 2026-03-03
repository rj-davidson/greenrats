"use client";

import type { GolferBio } from "@/features/picks/types";
import {
  CalendarIcon,
  GlobeIcon,
  GraduationCapIcon,
  MapPinIcon,
  RulerIcon,
  TrophyIcon,
  WeightIcon,
} from "lucide-react";
import type { ReactNode } from "react";

interface GolferDetailPanelProps {
  bio?: GolferBio | null;
  owgr?: number | null;
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

function InfoItem({ icon, label, value }: { icon: ReactNode; label: string; value: string }) {
  return (
    <div className="flex items-center gap-2 text-sm">
      <span className="text-muted-foreground">{icon}</span>
      <span className="text-muted-foreground">{label}</span>
      <span className="ml-auto tabular-nums sm:ml-0">{value}</span>
    </div>
  );
}

export function GolferDetailPanel({ bio, owgr }: GolferDetailPanelProps) {
  const hasAnyBio = bio != null && Object.values(bio).some((v) => v != null && v !== "");
  const hasOwgr = owgr != null && owgr > 0;

  if (!hasAnyBio && !hasOwgr) {
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
    <div className="grid grid-cols-1 gap-x-6 gap-y-2 p-4 sm:grid-cols-2 lg:grid-cols-3">
      {hasOwgr && (
        <InfoItem icon={<GlobeIcon className="size-4" />} label="OWGR" value={`#${owgr}`} />
      )}
      {birthDate && (
        <InfoItem icon={<CalendarIcon className="size-4" />} label="Born" value={birthDate} />
      )}
      {birthplace && (
        <InfoItem icon={<MapPinIcon className="size-4" />} label="From" value={birthplace} />
      )}
      {residence && (
        <InfoItem icon={<MapPinIcon className="size-4" />} label="Lives in" value={residence} />
      )}
      {bio?.school && (
        <InfoItem
          icon={<GraduationCapIcon className="size-4" />}
          label="School"
          value={bio.school}
        />
      )}
      {bio?.turned_pro && (
        <InfoItem
          icon={<TrophyIcon className="size-4" />}
          label="Turned Pro"
          value={String(bio.turned_pro)}
        />
      )}
      {bio?.height && (
        <InfoItem icon={<RulerIcon className="size-4" />} label="Height" value={bio.height} />
      )}
      {bio?.weight && (
        <InfoItem icon={<WeightIcon className="size-4" />} label="Weight" value={bio.weight} />
      )}
    </div>
  );
}
