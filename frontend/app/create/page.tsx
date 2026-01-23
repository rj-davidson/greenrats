"use client";

import { TopBar } from "@/components/core/top-bar";
import { Button } from "@/components/shadcn/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/shadcn/card";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/shadcn/form";
import { Input } from "@/components/shadcn/input";
import { useCreateLeague } from "@/features/leagues/queries";
import { createLeagueRequestSchema, type CreateLeagueRequest } from "@/features/leagues/types";
import { zodResolver } from "@hookform/resolvers/zod";
import { ArrowLeftIcon } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

export default function CreateLeaguePage() {
  const router = useRouter();
  const createLeague = useCreateLeague();

  const form = useForm<CreateLeagueRequest>({
    resolver: zodResolver(createLeagueRequestSchema),
    defaultValues: {
      name: "",
    },
  });

  const onSubmit = async (data: CreateLeagueRequest) => {
    try {
      const response = await createLeague.mutateAsync(data);
      toast.success(`Created ${response.league.name}!`);
      router.push(`/${response.league.id}`);
    } catch {
      toast.error("Failed to create league. Please try again.");
    }
  };

  return (
    <div className="flex min-h-screen flex-col">
      <TopBar />
      <main className="flex flex-1 items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Create a New League</CardTitle>
            <CardDescription>
              Create a league and invite your friends to join using the generated code.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                <FormField
                  control={form.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>League Name</FormLabel>
                      <FormControl>
                        <Input placeholder="Enter league name" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className="flex justify-between gap-2">
                  <Button type="button" variant="ghost" asChild>
                    <Link href="/">
                      <ArrowLeftIcon className="mr-2 size-4" />
                      Back
                    </Link>
                  </Button>
                  <Button type="submit" disabled={createLeague.isPending}>
                    {createLeague.isPending ? "Creating..." : "Create League"}
                  </Button>
                </div>
              </form>
            </Form>
          </CardContent>
        </Card>
      </main>
    </div>
  );
}
