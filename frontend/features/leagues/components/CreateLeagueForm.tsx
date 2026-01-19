"use client";

import { useCreateLeague } from "@/features/leagues/queries";
import { createLeagueRequestSchema, type CreateLeagueRequest } from "@/features/leagues/types";
import { Button } from "@/components/shadcn/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/shadcn/form";
import { Input } from "@/components/shadcn/input";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";

interface CreateLeagueFormProps {
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function CreateLeagueForm({ onSuccess, onCancel }: CreateLeagueFormProps) {
  const createLeague = useCreateLeague();

  const form = useForm<CreateLeagueRequest>({
    resolver: zodResolver(createLeagueRequestSchema),
    defaultValues: {
      name: "",
    },
  });

  const onSubmit = async (data: CreateLeagueRequest) => {
    try {
      await createLeague.mutateAsync(data);
      form.reset();
      onSuccess?.();
    } catch {
      // Error is handled by the mutation
    }
  };

  return (
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

        {createLeague.isError && (
          <p className="text-sm text-destructive">Failed to create league. Please try again.</p>
        )}

        <div className="flex justify-end gap-2">
          {onCancel && (
            <Button type="button" variant="outline" onClick={onCancel}>
              Cancel
            </Button>
          )}
          <Button type="submit" disabled={createLeague.isPending}>
            {createLeague.isPending ? "Creating..." : "Create League"}
          </Button>
        </div>
      </form>
    </Form>
  );
}
