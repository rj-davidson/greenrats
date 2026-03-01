import { Button } from "@/components/shadcn/button";
import { Separator } from "@/components/shadcn/separator";
import type { Metadata } from "next";
import Link from "next/link";

export const metadata: Metadata = {
  title: "How to Play",
  description:
    "Simple greenrats rules: pick one golfer per tournament, no repeats in your league, and finish with the highest season total in tournament earnings.",
};

const faqJsonLd = {
  "@context": "https://schema.org",
  "@type": "FAQPage",
  mainEntity: [
    {
      "@type": "Question",
      name: "How do greenrats fantasy golf leagues work?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Create a league and invite friends. Pick one golfer for each tournament. You can only use each golfer once per season in your league. Your score is the total tournament earnings from your picks.",
      },
    },
    {
      "@type": "Question",
      name: "Can I pick the same golfer twice?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "No. In each league, you can only use a golfer once per season.",
      },
    },
    {
      "@type": "Question",
      name: "What time zone are pick deadlines based on?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Pick deadlines use the tournament's local time zone, not your local time zone.",
      },
    },
  ],
};

function TheBasics() {
  return (
    <section className="space-y-4">
      <h2 className="text-2xl font-semibold">Quick Rules</h2>
      <ul className="space-y-3 text-muted-foreground">
        <li className="flex gap-3">
          <span className="font-semibold text-foreground">1.</span>
          <span>
            <strong className="text-foreground">Pick one golfer per tournament</strong> before the
            deadline.
          </span>
        </li>
        <li className="flex gap-3">
          <span className="font-semibold text-foreground">2.</span>
          <span>
            <strong className="text-foreground">No repeats in your league</strong> - once you use a
            golfer, you can&apos;t use them again that season.
          </span>
        </li>
        <li className="flex gap-3">
          <span className="font-semibold text-foreground">3.</span>
          <span>
            <strong className="text-foreground">Highest total earnings wins</strong> - your season
            score is the total prize money earned by your picks.
          </span>
        </li>
      </ul>
    </section>
  );
}

function MakingYourPick() {
  return (
    <section className="space-y-4">
      <h2 className="text-2xl font-semibold">Pick Deadlines</h2>
      <div className="space-y-3 text-muted-foreground">
        <p>
          The pick window <strong className="text-foreground">opens 4 days</strong> before each
          tournament and{" "}
          <strong className="text-foreground">
            closes at 5 AM in the tournament&apos;s local time zone
          </strong>{" "}
          on the morning the tournament begins.
        </p>
        <p>
          <strong className="text-foreground">Important:</strong> the deadline uses the
          tournament&apos;s local time zone (where the event is played), not your local time.
        </p>
        <p>
          You can <strong className="text-foreground">change your pick</strong> as often as you want
          until the deadline. After that, your pick is locked.
        </p>
        <p>
          <strong className="text-foreground">Miss the deadline?</strong> You get $0 for that
          tournament and can make a new pick for the next one.
        </p>
        <p>
          Simple strategy: balance recent form, course fit, and whether you want to save top golfers
          for later events.
        </p>
      </div>
    </section>
  );
}

function Scoring() {
  return (
    <section className="space-y-4">
      <h2 className="text-2xl font-semibold">How Scoring Works</h2>
      <div className="space-y-4 text-muted-foreground">
        <p>
          Your score for each tournament is the prize money your pick earns, and that amount is
          added to your season total.
        </p>

        <div className="rounded-lg border bg-muted/30 p-4">
          <p className="mb-3 text-sm font-medium text-foreground">Simple Example</p>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span>The Masters - Scottie Scheffler (Win)</span>
              <span className="font-medium text-foreground">$3,600,000</span>
            </div>
            <div className="flex justify-between">
              <span>US Open - Bryson DeChambeau (Win)</span>
              <span className="font-medium text-foreground">$4,300,000</span>
            </div>
            <div className="flex justify-between">
              <span>The Open - Brian Harman (Win)</span>
              <span className="font-medium text-foreground">$3,000,000</span>
            </div>
            <div className="flex justify-between">
              <span>PGA Championship - Collin Morikawa (MC)</span>
              <span className="font-medium text-foreground">$0</span>
            </div>
            <Separator className="my-2" />
            <div className="flex justify-between font-semibold text-foreground">
              <span>Your Total</span>
              <span>$10,900,000</span>
            </div>
          </div>
        </div>

        <p>
          <strong className="text-foreground">Missed cut?</strong> That pick scores $0 for the
          tournament.
        </p>
      </div>
    </section>
  );
}

function WinningYourLeague() {
  return (
    <section className="space-y-4">
      <h2 className="text-2xl font-semibold">How You Win</h2>
      <div className="space-y-3 text-muted-foreground">
        <p>
          This is a <strong className="text-foreground">season-long game</strong>. Every tournament
          adds to your total.
        </p>
        <p>
          Your league leaderboard shows everyone&apos;s cumulative earnings, so you can see where
          you stand after each event.
        </p>
        <p>
          The no-repeat rule is the main strategy decision: use top golfers now, or save them for
          better spots later.
        </p>
        <p>
          At the end of the season, the member with the{" "}
          <strong className="text-foreground">highest total earnings</strong> finishes first.
        </p>
      </div>
    </section>
  );
}

export default function RulesPage() {
  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(faqJsonLd) }}
      />
      <main className="flex min-h-screen flex-col items-center px-4 py-8 sm:px-8">
        <div className="w-full max-w-2xl">
          <div className="mb-12 text-center">
            <h1 className="mb-2 text-4xl font-bold">How to Play</h1>
            <p className="text-lg text-muted-foreground">
              Simple rules for picks, deadlines, and season scoring
            </p>
          </div>

          <div className="space-y-10">
            <TheBasics />
            <MakingYourPick />
            <Scoring />
            <WinningYourLeague />
          </div>

          <Separator className="my-10" />

          <div className="flex justify-center">
            <Link href="/login">
              <Button size="lg">Get Started</Button>
            </Link>
          </div>
        </div>
      </main>
    </>
  );
}
