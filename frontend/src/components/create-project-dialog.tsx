"use client";

import * as React from "react";
import { Plus } from "lucide-react";

import { createProjectAction } from "@/app/(app)/actions";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

export function CreateProjectDialog({
  open: openProp,
  onOpenChange,
  trigger,
}: {
  /** Controlled open state (omit for an internally-managed dialog). */
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  /** Custom trigger; pass null to render none (controlled from outside). */
  trigger?: React.ReactNode | null;
}) {
  const [openState, setOpenState] = React.useState(false);
  const open = openProp ?? openState;
  const setOpen = onOpenChange ?? setOpenState;

  const [name, setName] = React.useState("");
  const [error, setError] = React.useState<string | null>(null);
  const [pending, startTransition] = React.useTransition();

  function submit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    startTransition(async () => {
      const res = await createProjectAction(name);
      if (res.ok) {
        setName("");
        setOpen(false);
      } else {
        setError(res.error);
      }
    });
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      {trigger !== null && (
        <DialogTrigger asChild>
          {trigger ?? (
            <Button className="bg-[#f4f4f6] text-[#0a0a0b] hover:bg-white">
              <Plus className="size-4" />
              New project
            </Button>
          )}
        </DialogTrigger>
      )}
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create project</DialogTitle>
          <DialogDescription>
            Projects group secrets across dev, staging and prod environments.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={submit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <label className="text-sm font-medium text-foreground" htmlFor="project-name">
              Name
            </label>
            <Input
              id="project-name"
              autoFocus
              placeholder="payments-api"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="font-mono"
            />
          </div>
          {error && <p className="text-sm text-critical">{error}</p>}
          <DialogFooter>
            <Button
              type="button"
              variant="ghost"
              onClick={() => setOpen(false)}
              disabled={pending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={pending || !name.trim()}>
              {pending ? "Creating…" : "Create project"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
