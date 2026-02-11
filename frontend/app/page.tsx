import {
  ExamplePickHistoryCard,
  ExampleSeasonProgressCard,
  ExampleStandingsCard,
  ExampleYourStatsCard,
} from "@/app/landing";
import { LeaguePickerContent } from "@/app/league-picker";
import { TopBar } from "@/components/core/top-bar";
import { Alert, AlertDescription, AlertTitle } from "@/components/shadcn/alert";
import { Button } from "@/components/shadcn/button";
import { buildGetUserLeaguesQueryOptions } from "@/features/leagues/queries";
import type { User } from "@/features/users/types";
import { makeServerRequest } from "@/lib/query/server-requestor";
import { HydrationBoundary, QueryClient, dehydrate } from "@tanstack/react-query";
import { withAuth } from "@workos-inc/authkit-nextjs";
import { AlertTriangleIcon, CalendarCheckIcon, Rat, TrophyIcon, UsersIcon } from "lucide-react";
import Link from "next/link";
import { redirect } from "next/navigation";

const jsonLd = {
  "@context": "https://schema.org",
  "@type": "WebApplication",
  name: "greenrats",
  description:
    "Fantasy golf league manager - create pick'em leagues, track picks, and manage season-long standings",
  url: "https://greenrats.com",
  applicationCategory: "GameApplication",
  operatingSystem: "Web",
  offers: {
    "@type": "Offer",
    price: "0",
    priceCurrency: "USD",
  },
};

function HeroSection() {
  return (
    <section className="flex flex-col items-center py-12 text-center">
      <div className="mb-16 flex items-center justify-center gap-3">
        <div className="rounded-full bg-primary/10 p-3">
          <Rat className="size-10 text-primary" />
        </div>
        <h1 className="font-serif text-5xl tracking-wide sm:text-6xl">greenrats</h1>
      </div>

      <p className="mb-8 max-w-lg text-xl text-muted-foreground sm:text-2xl">
        A season-long golf pick&apos;em where every golfer can only be used once.
      </p>

      <div className="mb-10 flex flex-wrap justify-center gap-x-8 gap-y-3 text-sm text-muted-foreground">
        <span className="flex items-center gap-1.5">
          <CalendarCheckIcon className="size-4 text-primary" />
          One pick per tournament
        </span>
        <span className="flex items-center gap-1.5">
          <TrophyIcon className="size-4 text-primary" />
          Track league standings
        </span>
        <span className="flex items-center gap-1.5">
          <UsersIcon className="size-4 text-primary" />
          Compete in leagues
        </span>
      </div>

      <div className="flex gap-4">
        <Link href="/login">
          <Button size="lg">Get Started</Button>
        </Link>
        <Link href="/rules">
          <Button variant="outline" size="lg">
            How to Play
          </Button>
        </Link>
      </div>

      <p className="mt-6 text-sm text-muted-foreground">
        Already have an account?{" "}
        <Link href="/login" className="text-primary underline-offset-4 hover:underline">
          Sign in
        </Link>
      </p>
    </section>
  );
}

function ShowcaseSection() {
  return (
    <section className="w-full max-w-5xl py-12">
      <h2 className="mb-8 text-center text-2xl font-semibold">See It In Action</h2>
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        <ExampleStandingsCard />
        <ExampleYourStatsCard />
        <ExamplePickHistoryCard />
      </div>
      <div className="mt-6">
        <ExampleSeasonProgressCard />
      </div>
    </section>
  );
}

function HowItWorksSection() {
  const steps = [
    {
      icon: <CalendarCheckIcon className="size-8" />,
      title: "Pick Your Golfer",
      description:
        "Choose one golfer before each tournament begins. Use them wisely - no repeats allowed.",
    },
    {
      icon: <TrophyIcon className="size-8" />,
      title: "Track Standings",
      description:
        "League standings are based on your picks' tournament prize money. Watch your total climb all season.",
    },
    {
      icon: <UsersIcon className="size-8" />,
      title: "Compete in Leagues",
      description: "Create or join leagues with friends. The highest total at season end wins.",
    },
  ];

  return (
    <section className="w-full max-w-4xl py-12">
      <h2 className="mb-8 text-center text-2xl font-semibold">How It Works</h2>
      <div className="grid gap-8 md:grid-cols-3">
        {steps.map((step) => (
          <div key={step.title} className="flex flex-col items-center text-center">
            <div className="mb-4 rounded-full bg-primary/10 p-4 text-primary">{step.icon}</div>
            <h3 className="mb-2 text-lg font-medium">{step.title}</h3>
            <p className="text-sm text-muted-foreground">{step.description}</p>
          </div>
        ))}
      </div>
    </section>
  );
}

function CtaSection() {
  return (
    <section className="w-full max-w-2xl py-16 text-center">
      <h2 className="mb-4 text-2xl font-semibold">Ready to play?</h2>
      <p className="mb-8 text-muted-foreground">
        Join now and start making picks for this week&apos;s tournament.
      </p>
      <div className="flex justify-center gap-4">
        <Link href="/login">
          <Button size="lg">Get Started</Button>
        </Link>
        <Link href="/rules">
          <Button variant="outline" size="lg">
            Learn the Rules
          </Button>
        </Link>
      </div>
    </section>
  );
}

function ServiceOutageAlert() {
  return (
    <section className="w-full max-w-5xl">
      <Alert className="border-amber-300 bg-amber-50 text-amber-900">
        <AlertTriangleIcon className="size-4" />
        <AlertTitle>Temporary sign-in outage</AlertTitle>
        <AlertDescription className="text-amber-900/90">
          One of our service providers is currently experiencing an outage that is preventing sign
          in to the app. Sorry about this, and please try again later today.
        </AlertDescription>
      </Alert>
    </section>
  );
}

function LandingPage() {
  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      <main className="flex min-h-screen flex-col items-center px-4 py-8 sm:px-8">
        <ServiceOutageAlert />
        <HeroSection />
        <ShowcaseSection />
        <HowItWorksSection />
        <CtaSection />
      </main>
    </>
  );
}

interface LeaguePickerProps {
  displayName: string;
  dehydratedState: ReturnType<typeof dehydrate>;
}

function LeaguePicker({ displayName, dehydratedState }: LeaguePickerProps) {
  return (
    <div className="flex min-h-screen flex-col">
      <TopBar />
      <main className="flex-1 p-8">
        <HydrationBoundary state={dehydratedState}>
          <LeaguePickerContent displayName={displayName} />
        </HydrationBoundary>
      </main>
    </div>
  );
}

export default async function Home() {
  const { user: authUser } = await withAuth();

  if (!authUser) {
    return <LandingPage />;
  }

  let user: User;
  try {
    user = await makeServerRequest.get<User>("/api/v1/users/me");
  } catch {
    return <LandingPage />;
  }

  if (!user.display_name) {
    redirect("/onboarding");
  }

  const queryClient = new QueryClient();
  try {
    await queryClient.prefetchQuery(buildGetUserLeaguesQueryOptions(makeServerRequest));
  } catch {}

  return (
    <LeaguePicker
      displayName={user.display_name || authUser.email || "User"}
      dehydratedState={dehydrate(queryClient)}
    />
  );
}
