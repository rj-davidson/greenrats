"use client";

import { InputGroup, InputGroupButton, InputGroupInput } from "@/components/shadcn/input-group";
import { useJoinLeague } from "@/features/leagues/queries";
import { ArrowRight } from "lucide-react";
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
          router.push(`/${data.league.id}`);
        },
        onError: (error) => {
          const message = error instanceof Error ? error.message : "Failed to join league";
          toast.error(message);
        },
      },
    );
  };

  return (
    <form onSubmit={handleSubmit}>
      <InputGroup className="w-fit">
        <InputGroupInput
          type="text"
          placeholder="League code"
          value={code}
          onChange={(e) => setCode(e.target.value.toUpperCase())}
          className="w-28"
          maxLength={6}
        />
        <InputGroupButton type="submit" disabled={!code.trim() || joinLeague.isPending}>
          <ArrowRight className="size-4" />
        </InputGroupButton>
      </InputGroup>
    </form>
  );
}
