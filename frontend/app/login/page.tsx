import { Button } from "@/components/shadcn/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/shadcn/card";
import { getSignInUrl } from "@workos-inc/authkit-nextjs";
import { redirect } from "next/navigation";

export default async function LoginPage() {
  const signInUrl = await getSignInUrl();

  async function signIn() {
    "use server";
    const url = await getSignInUrl();
    redirect(url);
  }

  return (
    <main className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Welcome to GreenRats</CardTitle>
          <CardDescription>Sign in to start making your picks</CardDescription>
        </CardHeader>
        <CardContent>
          <form action={signIn}>
            <Button type="submit" className="w-full" size="lg">
              Sign in with WorkOS
            </Button>
          </form>
        </CardContent>
      </Card>
    </main>
  );
}
