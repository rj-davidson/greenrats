import { TopBar } from "@/components/core/top-bar";
import { Button } from "@/components/shadcn/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/shadcn/card";
import { OnboardingForm } from "@/features/users/components";
import type { User } from "@/features/users/types";
import { makeServerRequest } from "@/lib/query/server-requestor";
import { getSignInUrl, withAuth } from "@workos-inc/authkit-nextjs";
import { redirect } from "next/navigation";

export default async function OnboardingPage() {
  async function signIn() {
    "use server";
    const url = await getSignInUrl();
    redirect(url);
  }

  const { user: authUser } = await withAuth();

  if (!authUser) {
    return (
      <div className="flex min-h-screen flex-col">
        <TopBar />
        <main className="flex flex-1 items-center justify-center p-4">
          <Card className="w-full max-w-md">
            <CardHeader className="text-center">
              <CardTitle className="text-2xl">Finish onboarding</CardTitle>
              <CardDescription>Sign in to set your display name.</CardDescription>
            </CardHeader>
            <CardContent>
              <form action={signIn}>
                <Button type="submit" className="w-full" size="lg">
                  Sign in
                </Button>
              </form>
            </CardContent>
          </Card>
        </main>
      </div>
    );
  }

  // Fetch the current user
  let user: User | null = null;
  try {
    user = await makeServerRequest.get<User>("/api/v1/users/me");
  } catch {
    // User not authenticated, redirect to login
    redirect("/login");
  }

  if (user.display_name) {
    redirect("/");
  }

  return (
    <div className="flex min-h-screen flex-col">
      <TopBar />
      <main className="flex flex-1 items-center justify-center p-4">
        <OnboardingForm />
      </main>
    </div>
  );
}
