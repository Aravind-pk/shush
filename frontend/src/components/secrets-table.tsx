"use client";

import * as React from "react";
import { Check, Copy, Eye, EyeOff, Loader2 } from "lucide-react";

import type { Secret, Environment } from "@/lib/shush";
import { getSecretAction } from "@/app/(app)/actions";
import { cn } from "@/lib/utils";
import { AddSecretDialog } from "@/components/add-secret-dialog";

export function SecretsTable({
  secrets,
  projectId,
  environment,
}: {
  secrets: Secret[];
  projectId: string;
  environment: Environment;
}) {
  return (
    <div className="overflow-hidden rounded-[14px] border border-white/[0.07] bg-[linear-gradient(180deg,#131315,#0f0f11)]">
      {/* Header row */}
      <div className="grid grid-cols-[1.6fr_2fr_0.5fr_0.7fr] border-b border-white/[0.07] px-5 py-3 text-[11.5px] uppercase tracking-[0.04em] text-[#6a6a70]">
        <span>Key</span>
        <span>Value</span>
        <span>Ver</span>
        <span className="text-right">Actions</span>
      </div>

      {secrets.length === 0 ? (
        <div className="px-5 py-12 text-center text-sm text-[#7d7d84]">
          No secrets in{" "}
          <span className="font-mono text-[#cfcfd5]">{environment}</span> yet.
        </div>
      ) : (
        secrets.map((s) => (
          // Keying by version resets the cached plaintext when a secret is
          // edited, so a stale revealed value can't linger after an update.
          <SecretRow
            key={`${s.key}:${s.version ?? 1}`}
            secret={s}
            projectId={projectId}
            environment={environment}
          />
        ))
      )}

      <AddSecretDialog
        projectId={projectId}
        environment={environment}
        trigger={
          <button className="flex w-full items-center gap-2 border-t border-white/[0.05] px-5 py-[13px] text-left text-[13.5px] text-[#7b9ef0] transition-colors hover:bg-white/[0.02]">
            <svg
              className="size-4"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M12 5v14M5 12h14" />
            </svg>
            Add secret
          </button>
        }
      />
    </div>
  );
}

function SecretRow({
  secret,
  projectId,
  environment,
}: {
  secret: Secret;
  projectId: string;
  environment: Environment;
}) {
  // value === null means the plaintext hasn't been fetched yet. It's only
  // loaded on reveal/copy, via the backend's GetSecret RPC.
  const [value, setValue] = React.useState<string | null>(null);
  const [revealed, setRevealed] = React.useState(false);
  const [loading, setLoading] = React.useState(false);
  const [copied, setCopied] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);

  // Returns the plaintext, fetching it once and caching for this row.
  async function ensureValue(): Promise<string | null> {
    if (value !== null) return value;
    setLoading(true);
    setError(null);
    const res = await getSecretAction(projectId, environment, secret.key);
    setLoading(false);
    if (res.ok) {
      setValue(res.value);
      return res.value;
    }
    setError(res.error);
    return null;
  }

  async function toggleReveal() {
    if (revealed) {
      setRevealed(false);
      return;
    }
    const v = await ensureValue();
    if (v !== null) setRevealed(true);
  }

  async function copy() {
    const v = await ensureValue();
    if (v === null) return;
    try {
      await navigator.clipboard.writeText(v);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard may be unavailable (e.g. non-secure context); ignore.
    }
  }

  return (
    <div className="grid grid-cols-[1.6fr_2fr_0.5fr_0.7fr] items-center border-b border-white/[0.03] px-5 transition-colors hover:bg-white/[0.02]">
      <span className="truncate font-mono text-[13px] text-[#dcdce0]">
        {secret.key}
      </span>
      <span
        className={cn(
          "truncate py-3.5 font-mono text-[13px]",
          error
            ? "text-critical"
            : revealed
              ? "text-[#dcdce0]"
              : "tracking-[1px] text-[#6a6a70]",
        )}
      >
        {error ? (
          error
        ) : loading ? (
          <span className="inline-flex items-center gap-1.5 text-[#6a6a70]">
            <Loader2 className="size-3.5 animate-spin" />
            Decrypting…
          </span>
        ) : revealed ? (
          value || <span className="text-[#6a6a70] italic">(empty)</span>
        ) : (
          "••••••••••••••"
        )}
      </span>
      <span>
        <span className="rounded-[5px] border border-white/[0.08] bg-[#1c1c20] px-1.5 py-0.5 font-mono text-[11px] text-[#9a9aa0]">
          v{secret.version ?? 1}
        </span>
      </span>
      <div className="flex items-center justify-end gap-3 text-[#5f5f66]">
        <button
          onClick={toggleReveal}
          disabled={loading}
          title={revealed ? "Hide" : "Reveal"}
          className="transition-colors hover:text-[#cfcfd5] disabled:opacity-50"
        >
          {revealed ? <EyeOff className="size-[15px]" /> : <Eye className="size-[15px]" />}
        </button>
        <button
          onClick={copy}
          disabled={loading}
          title="Copy value"
          className={cn(
            "transition-colors hover:text-[#cfcfd5] disabled:opacity-50",
            copied && "text-success",
          )}
        >
          {copied ? <Check className="size-[15px]" /> : <Copy className="size-[15px]" />}
        </button>
      </div>
    </div>
  );
}
