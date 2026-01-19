"use client";

import { useRouter } from "next/navigation";
import { useJoinLeague } from "../queries";
import { joinLeagueRequestSchema } from "../types";
import { Button } from "@/components/shadcn/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/shadcn/card";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/shadcn/form";
import { Input } from "@/components/shadcn/input";
import { zodResolver } from "@hookform/resolvers/zod";
import { UsersIcon } from "lucide-react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import type { z } from "zod";

type FormValues = z.infer<typeof joinLeagueRequestSchema>;

export function JoinLeagueForm() {
  const router = useRouter();
  const joinLeague = useJoinLeague();

  const form = useForm<FormValues>({
    resolver: zodResolver(joinLeagueRequestSchema),
    defaultValues: {
      code: "",
    },
  });

  const onSubmit = async (data: FormValues) => {
    try {
      const result = await joinLeague.mutateAsync(data);
      toast.success(`Joined ${result.league.name}!`);
      router.push(`/leagues/${result.league.id}`);
    } catch (error) {
      const message = error instanceof Error ? error.message : "Failed to join league";
      if (message.includes("invalid join code")) {
        form.setError("code", { message: "Invalid join code" });
      } else if (message.includes("already a member")) {
        toast.error("You are already a member of this league");
      } else if (message.includes("joining is disabled")) {
        toast.error("This league is not accepting new members");
      } else if (message.includes("maximum members")) {
        toast.error("This league has reached its member limit");
      } else {
        toast.error(message);
      }
    }
  };

  return (
    <Card className="mx-auto max-w-md">
      <CardHeader className="text-center">
        <div className="bg-primary/10 mx-auto mb-2 w-fit rounded-full p-3">
          <UsersIcon className="text-primary size-6" />
        </div>
        <CardTitle>Join a League</CardTitle>
        <CardDescription>Enter the join code to become a member</CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="code"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Join Code</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="ABC123"
                      {...field}
                      className="text-center font-mono text-lg uppercase tracking-widest"
                      onChange={(e) => field.onChange(e.target.value.toUpperCase())}
                    />
                  </FormControl>
                  <FormDescription>Ask your league commissioner for the code</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <Button type="submit" className="w-full" disabled={joinLeague.isPending}>
              {joinLeague.isPending ? "Joining..." : "Join League"}
            </Button>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
