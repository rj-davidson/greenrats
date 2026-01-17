import { Button } from "@/components/shadcn/button";
import { withAuth, getSignInUrl } from "@workos-inc/authkit-nextjs";
import Link from "next/link";
import { redirect } from "next/navigation";

export default async function Home() {
  const { user } = await withAuth();

  if (user) {
    redirect("/dashboard");
  }

  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-8">
      <div className="max-w-2xl text-center">
        <h1 className="mb-4 text-5xl font-bold">GreenRats</h1>
        <p className="text-muted-foreground mb-8 text-xl">
          Pick one golfer per tournament. Compete with friends. Track your earnings throughout the
          PGA Tour season.
        </p>
        <div className="flex justify-center gap-4">
          <Link href="/login">
            <Button size="lg">Get Started</Button>
          </Link>
        </div>
      </div>
    </main>
  );
}
