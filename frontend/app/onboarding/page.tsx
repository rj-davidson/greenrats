import { OnboardingForm } from "@/features/users/components";
import type { User } from "@/features/users/types";
import { makeServerRequest } from "@/lib/query/server-requestor";
import { redirect } from "next/navigation";

export default async function OnboardingPage() {
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
    <main className="flex min-h-screen items-center justify-center p-4">
      <OnboardingForm />
    </main>
  );
}
