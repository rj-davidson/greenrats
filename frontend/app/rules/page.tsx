import { Button } from "@/components/shadcn/button";
import { Separator } from "@/components/shadcn/separator";
import Link from "next/link";

function TheBasics() {
  return (
    <section className="space-y-4">
      <h2 className="text-2xl font-semibold">The Basics</h2>
      <ul className="space-y-3 text-muted-foreground">
        <li className="flex gap-3">
          <span className="font-semibold text-foreground">1.</span>
          <span>
            <strong className="text-foreground">Pick one golfer per tournament</strong> - before
            each event, you select one golfer to represent you.
          </span>
        </li>
        <li className="flex gap-3">
          <span className="font-semibold text-foreground">2.</span>
          <span>
            <strong className="text-foreground">No repeats</strong> - once you pick a golfer, you
            can&apos;t use them again for the rest of the season. Choose wisely.
          </span>
        </li>
        <li className="flex gap-3">
          <span className="font-semibold text-foreground">3.</span>
          <span>
            <strong className="text-foreground">Highest total earnings wins</strong> - your score is
            the sum of prize money from all your picks throughout the season.
          </span>
        </li>
      </ul>
    </section>
  );
}

function MakingYourPick() {
  return (
    <section className="space-y-4">
      <h2 className="text-2xl font-semibold">Making Your Pick</h2>
      <div className="space-y-3 text-muted-foreground">
        <p>
          The pick window <strong className="text-foreground">opens 4 days</strong> before each
          tournament and{" "}
          <strong className="text-foreground">closes at 5 AM local time</strong> the morning the
          tournament begins.
        </p>
        <p>
          While the window is open, you can{" "}
          <strong className="text-foreground">change your mind</strong> as many times as you want.
          Once the window closes, your pick is locked in.
        </p>
        <p>
          <strong className="text-foreground">Forgot to pick?</strong> If you miss the window, you
          simply don&apos;t earn anything for that tournament. It happens - just get the next one.
        </p>
        <p>
          When deciding, consider who&apos;s playing well lately, who fits the course, and whether
          you want to save a top golfer for a major later in the season.
        </p>
      </div>
    </section>
  );
}

function Scoring() {
  return (
    <section className="space-y-4">
      <h2 className="text-2xl font-semibold">Scoring</h2>
      <div className="space-y-4 text-muted-foreground">
        <p>
          Your pick earns whatever prize money they win that week. First place takes home the
          biggest check, but even making the cut adds to your total.
        </p>

        <div className="rounded-lg border bg-muted/30 p-4">
          <p className="mb-3 text-sm font-medium text-foreground">Example Season</p>
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
          <strong className="text-foreground">Missed cut?</strong> If your golfer doesn&apos;t make
          the weekend, they earn $0 for that tournament. That&apos;s golf - it happens to everyone.
        </p>
      </div>
    </section>
  );
}

function WinningYourLeague() {
  return (
    <section className="space-y-4">
      <h2 className="text-2xl font-semibold">Winning Your League</h2>
      <div className="space-y-3 text-muted-foreground">
        <p>
          This is a <strong className="text-foreground">season-long competition</strong>. Every
          tournament matters, from the first event in January through the Tour Championship in
          August.
        </p>
        <p>
          Your league leaderboard tracks everyone&apos;s cumulative earnings. You can see who&apos;s
          leading, who&apos;s behind, and how your picks compare to your friends.
        </p>
        <p>
          The no-repeat rule is what makes this interesting. Do you use Scottie Scheffler at Augusta
          where he dominates? Or save him for a different major where fewer people will pick him?
        </p>
        <p>
          At season&apos;s end, the player with the{" "}
          <strong className="text-foreground">highest total earnings</strong> wins bragging rights
          for the year.
        </p>
      </div>
    </section>
  );
}

export default function RulesPage() {
  return (
    <main className="flex min-h-screen flex-col items-center px-4 py-8 sm:px-8">
      <div className="w-full max-w-2xl">
        <div className="mb-12 text-center">
          <h1 className="mb-2 text-4xl font-bold">How to Play</h1>
          <p className="text-lg text-muted-foreground">
            Everything you need to know to get started
          </p>
        </div>

        <div className="space-y-10">
          <TheBasics />
          <MakingYourPick />
          <Scoring />
          <WinningYourLeague />
        </div>

        <Separator className="my-10" />

        <div className="flex justify-center gap-4">
          <Link href="/login">
            <Button size="lg">Get Started</Button>
          </Link>
          <Link href="/">
            <Button variant="outline" size="lg">
              Back to Home
            </Button>
          </Link>
        </div>
      </div>
    </main>
  );
}
