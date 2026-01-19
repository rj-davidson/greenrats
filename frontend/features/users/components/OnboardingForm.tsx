"use client";

import { useCheckDisplayName, useSetDisplayName } from "@/features/users/queries";
import { type SetDisplayNameRequest, setDisplayNameRequestSchema } from "@/features/users/types";
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
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/shadcn/form";
import { Input } from "@/components/shadcn/input";
import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";

export function OnboardingForm() {
  const router = useRouter();
  const [debouncedName, setDebouncedName] = useState("");

  const form = useForm<SetDisplayNameRequest>({
    resolver: zodResolver(setDisplayNameRequestSchema),
    defaultValues: {
      display_name: "",
    },
  });

  const watchedName = form.watch("display_name");

  // Debounce the name for availability check
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedName(watchedName);
    }, 300);
    return () => clearTimeout(timer);
  }, [watchedName]);

  const { data: availabilityData, isLoading: isCheckingAvailability } =
    useCheckDisplayName(debouncedName);

  const setDisplayName = useSetDisplayName();

  const isAvailable = availabilityData?.available ?? true;
  const showAvailability = debouncedName.length >= 3 && !isCheckingAvailability;

  async function onSubmit(data: SetDisplayNameRequest) {
    try {
      await setDisplayName.mutateAsync(data);
      router.push("/");
    } catch {
      form.setError("display_name", {
        type: "manual",
        message: "Failed to set display name. Please try again.",
      });
    }
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader className="text-center">
        <CardTitle className="text-2xl">Choose Your Display Name</CardTitle>
        <CardDescription>
          This is how other players will see you on the leaderboard. Choose wisely - you won&apos;t
          be able to change it later!
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
            <FormField
              control={form.control}
              name="display_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Display Name</FormLabel>
                  <FormControl>
                    <Input placeholder="Enter a unique display name" {...field} />
                  </FormControl>
                  <FormDescription>
                    <span className="mb-1 block text-muted-foreground">
                      3-20 characters. Letters, numbers, and underscores only.
                    </span>
                    {showAvailability && (
                      <span className={isAvailable ? "text-green-600" : "text-red-600"}>
                        {isAvailable ? "This name is available!" : "This name is already taken."}
                      </span>
                    )}
                    {isCheckingAvailability && debouncedName.length >= 3 && (
                      <span className="text-muted-foreground">Checking availability...</span>
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <Button
              type="submit"
              className="w-full"
              disabled={
                !isAvailable ||
                setDisplayName.isPending ||
                watchedName.length < 3 ||
                watchedName.length > 20
              }
            >
              {setDisplayName.isPending ? "Saving..." : "Continue"}
            </Button>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
