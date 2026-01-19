import { Button } from "@/components/shadcn/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/shadcn/card";
import { Separator } from "@/components/shadcn/separator";
import Link from "next/link";

export default function RulesPage() {
  return (
    <main className="flex min-h-screen flex-col items-center p-8">
      <div className="w-full max-w-3xl">
        <div className="mb-8 text-center">
          <h1 className="mb-2 text-4xl font-bold">How to Play</h1>
          <p className="text-muted-foreground text-lg">
            Learn the rules of GreenRats golf pick&apos;em
          </p>
        </div>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Pick Uniqueness</CardTitle>
              <CardDescription>
                Strategic allocation across the season
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <p>
                Each tournament, you pick <strong>one golfer</strong> to
                represent you.
              </p>
              <p>
                Once you pick a golfer, you{" "}
                <strong>cannot use them again</strong> for the rest of the
                season within your league.
              </p>
              <p>
                With 40+ tournaments in a PGA Tour season, you&apos;ll need to
                think strategically about when to use your top picks on major
                events versus saving them for later.
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Leaderboard Scoring</CardTitle>
              <CardDescription>
                Combined prize money earnings
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <p>
                Your score is the <strong>total prize money</strong> earned by
                all your picks throughout the season.
              </p>
              <p>
                For example: if your pick wins $2,000,000 at the Masters and
                your next pick earns $500,000 at the Open Championship, your
                total score would be <strong>$2,500,000</strong>.
              </p>
              <p>
                The player with the highest cumulative earnings at the end of
                the season wins.
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Pick Timing</CardTitle>
              <CardDescription>When you can make and change picks</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <p>
                The pick window <strong>opens 3 days</strong> before each
                tournament begins.
              </p>
              <p>
                The window <strong>closes</strong> when the tournament starts
                (first tee time on Thursday).
              </p>
              <p>
                While the window is open, you can change your pick as many times
                as you want. Once the tournament starts, your pick is locked in.
              </p>
            </CardContent>
          </Card>
        </div>

        <Separator className="my-8" />

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
