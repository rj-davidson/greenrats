"use client";

import { Button } from "@/components/shadcn/button";
import { Input } from "@/components/shadcn/input";
import { useJoinLeague } from "@/features/leagues/queries";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { toast } from "sonner";

export function QuickJoinInput() {
  const [code, setCode] = useState("");
  const router = useRouter();
  const joinLeague = useJoinLeague();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const trimmedCode = code.trim().toUpperCase();
    if (!trimmedCode) return;

    joinLeague.mutate(
      { code: trimmedCode },
      {
        onSuccess: (data) => {
          toast.success(`Joined ${data.league.name}!`);
          setCode("");
          router.push(`/leagues/${data.league.id}`);
        },
        onError: (error) => {
          const message = error instanceof Error ? error.message : "Failed to join league";
          toast.error(message);
        },
      },
    );
  };

  return (
    <form onSubmit={handleSubmit} className="flex items-center gap-2">
      <Input
        type="text"
        placeholder="Enter code"
        value={code}
        onChange={(e) => setCode(e.target.value.toUpperCase())}
        className="w-28 font-mono uppercase"
        maxLength={6}
      />
      <Button type="submit" size="sm" disabled={!code.trim() || joinLeague.isPending}>
        {joinLeague.isPending ? "Joining..." : "Join"}
      </Button>
    </form>
  );
}
