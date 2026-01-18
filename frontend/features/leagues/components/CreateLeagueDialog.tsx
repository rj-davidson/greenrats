"use client";

import { CreateLeagueForm } from "./CreateLeagueForm";
import { Button } from "@/components/shadcn/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/shadcn/dialog";
import { PlusIcon } from "lucide-react";
import { useState } from "react";

export function CreateLeagueDialog() {
  const [open, setOpen] = useState(false);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <PlusIcon className="mr-2 h-4 w-4" />
          Create League
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create a New League</DialogTitle>
          <DialogDescription>
            Create a league and invite your friends to join using the generated code.
          </DialogDescription>
        </DialogHeader>
        <CreateLeagueForm onSuccess={() => setOpen(false)} onCancel={() => setOpen(false)} />
      </DialogContent>
    </Dialog>
  );
}
