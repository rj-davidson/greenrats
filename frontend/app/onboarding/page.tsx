import { redirect } from "next/navigation";

import { OnboardingForm } from "@/features/users/components";
import type { User } from "@/features/users/types";
import { makeServerRequest } from "@/lib/query/server-requestor";

export default async function OnboardingPage() {
  // Fetch the current user
  let user: User | null = null;
  try {
    user = await makeServerRequest.get<User>("/api/v1/users/me");
  } catch {
    // User not authenticated, redirect to login
    redirect("/login");
  }

  // If user already has a display name, redirect to dashboard
  if (user.display_name) {
    redirect("/dashboard");
  }

  return (
    <main className="flex min-h-screen items-center justify-center p-4">
      <OnboardingForm />
    </main>
  );
}
