"use client";

import * as React from "react";
import { Plus } from "lucide-react";

import { putSecretAction } from "@/app/(app)/actions";
import type { Environment } from "@/lib/shush";
import { ENV_META } from "@/lib/env-meta";
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

export function AddSecretDialog({
  projectId,
  environment,
  trigger,
  open: openProp,
  onOpenChange,
}: {
  projectId: string;
  environment: Environment;
  /** Custom trigger; pass null to render none (controlled from outside). */
  trigger?: React.ReactNode | null;
  /** Controlled open state (omit for an internally-managed dialog). */
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}) {
  const [openState, setOpenState] = React.useState(false);
  const open = openProp ?? openState;
  const setOpen = onOpenChange ?? setOpenState;
  const [key, setKey] = React.useState("");
  const [value, setValue] = React.useState("");
  const [error, setError] = React.useState<string | null>(null);
  const [pending, startTransition] = React.useTransition();

  function submit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    startTransition(async () => {
      const res = await putSecretAction(projectId, environment, key, value);
      if (res.ok) {
        setKey("");
        setValue("");
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
              Add secret
            </Button>
          )}
        </DialogTrigger>
      )}
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add secret</DialogTitle>
          <DialogDescription className="flex items-center gap-1.5">
            Stored encrypted (AES-256-GCM) in
            <span className="inline-flex items-center gap-1 font-mono text-foreground">
              <span
                className="size-1.5 rounded-full"
                style={{ background: ENV_META[environment].dot }}
              />
              {environment}
            </span>
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={submit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <label className="text-sm font-medium text-foreground" htmlFor="secret-key">
              Key
            </label>
            <Input
              id="secret-key"
              autoFocus
              placeholder="STRIPE_API_KEY"
              value={key}
              onChange={(e) => setKey(e.target.value)}
              className="font-mono"
            />
          </div>
          <div className="flex flex-col gap-2">
            <label className="text-sm font-medium text-foreground" htmlFor="secret-value">
              Value
            </label>
            <Input
              id="secret-value"
              placeholder="sk_live_…"
              value={value}
              onChange={(e) => setValue(e.target.value)}
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
            <Button type="submit" disabled={pending || !key.trim()}>
              {pending ? "Saving…" : "Save secret"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
