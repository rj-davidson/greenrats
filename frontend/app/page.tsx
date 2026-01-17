import { withAuth, getSignInUrl } from "@workos-inc/authkit-nextjs";
import { redirect } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/shadcn/button";

export default async function Home() {
  const { user } = await withAuth();

  if (user) {
    redirect("/dashboard");
  }

  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-8">
      <div className="text-center max-w-2xl">
        <h1 className="text-5xl font-bold mb-4">GreenRats</h1>
        <p className="text-xl text-muted-foreground mb-8">
          Pick one golfer per tournament. Compete with friends. Track your
          earnings throughout the PGA Tour season.
        </p>
        <div className="flex gap-4 justify-center">
          <Link href="/login">
            <Button size="lg">Get Started</Button>
          </Link>
        </div>
      </div>
    </main>
  );
}
