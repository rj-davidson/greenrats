import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Pricing",
  description:
    "greenrats pricing plans for fantasy golf leagues - free tier available, premium features for serious league commissioners.",
};

export default function PricingPage() {
  return (
    <main className="mx-auto w-full max-w-3xl px-6 py-12">
      <h1 className="text-3xl font-semibold">Pricing</h1>
      <p className="mt-4 text-muted-foreground">Pricing information coming soon.</p>
    </main>
  );
}
