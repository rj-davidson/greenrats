import { Button } from "@/components/shadcn/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/shadcn/card";
import { getSignInUrl } from "@workos-inc/authkit-nextjs";
import { Rat } from "lucide-react";
import { redirect } from "next/navigation";

export default async function LoginPage() {
  async function signIn() {
    "use server";
    const url = await getSignInUrl();
    redirect(url);
  }

  return (
    <main className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="flex items-center justify-center gap-2 text-2xl">
            Welcome to <span className="font-serif tracking-wide">greenrats</span>{" "}
            <Rat className="size-6 text-primary" />
          </CardTitle>
          <CardDescription>Sign in to start making your picks</CardDescription>
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
  );
}
