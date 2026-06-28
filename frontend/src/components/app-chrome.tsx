"use client";

import * as React from "react";
import { Bell, Search, type LucideIcon } from "lucide-react";

import { cn } from "@/lib/utils";

// A breadcrumb segment. `mono` renders it in JetBrains Mono (used for
// project/resource names); the final segment is styled as the current page.
export type Crumb = { label: string; mono?: boolean };

// The page-specific primary action shown at the top-right of the top bar
// (e.g. "New project", "Add secret"). onClick typically opens a dialog.
export type ChromeAction = {
  label: string;
  icon: LucideIcon;
  onClick: () => void;
};

type ChromeState = {
  breadcrumb: Crumb[];
  action: ChromeAction | null;
};

const DEFAULT: ChromeState = { breadcrumb: [{ label: "Shush" }], action: null };

const ChromeContext = React.createContext<{
  state: ChromeState;
  set: (s: ChromeState) => void;
} | null>(null);

export function AppChromeProvider({ children }: { children: React.ReactNode }) {
  const [state, set] = React.useState<ChromeState>(DEFAULT);
  const value = React.useMemo(() => ({ state, set }), [state]);
  return (
    <ChromeContext.Provider value={value}>{children}</ChromeContext.Provider>
  );
}

function useChrome() {
  const ctx = React.useContext(ChromeContext);
  if (!ctx) throw new Error("useChrome must be used within AppChromeProvider");
  return ctx;
}

/**
 * Called by a page (via a small client component) to populate the top bar's
 * breadcrumb and primary action. Pass memoized values so the effect is stable.
 */
export function useSetChrome(breadcrumb: Crumb[], action: ChromeAction | null) {
  const { set } = useChrome();
  React.useEffect(() => {
    set({ breadcrumb, action });
    return () => set(DEFAULT);
  }, [set, breadcrumb, action]);
}

export function AppTopbar() {
  const { state } = useChrome();
  const { breadcrumb, action } = state;
  const ActionIcon = action?.icon;

  return (
    <div className="sticky top-0 z-30 flex items-center justify-between border-b border-white/[0.06] bg-[#0a0a0b]/85 px-[30px] py-4 backdrop-blur">
      {/* Breadcrumb */}
      <div className="flex items-center gap-[9px] text-[14.5px]">
        {breadcrumb.map((c, i) => {
          const last = i === breadcrumb.length - 1;
          return (
            <React.Fragment key={`${c.label}-${i}`}>
              <span
                className={cn(
                  last
                    ? "text-[#e9e9ec]"
                    : c.mono
                      ? "font-mono text-[#9a9aa0]"
                      : "text-[#7d7d84]",
                )}
              >
                {c.label}
              </span>
              {!last && <span className="text-[#3f3f45]">/</span>}
            </React.Fragment>
          );
        })}
      </div>

      {/* Search + notifications + primary action */}
      <div className="flex items-center gap-[10px]">
        <div className="flex h-9 w-[230px] items-center gap-[9px] rounded-[10px] border border-white/[0.09] bg-[#121214] px-3">
          <Search className="size-[15px] text-[#6a6a70]" />
          <span className="flex-1 text-[13px] text-[#6a6a70]">Search…</span>
          <span className="rounded-[5px] border border-white/[0.08] bg-[#1c1c20] px-1.5 py-px font-mono text-[11px] text-[#6a6a70]">
            ⌘K
          </span>
        </div>
        <button className="relative flex size-9 items-center justify-center rounded-[10px] border border-white/[0.09] bg-[#121214] text-[#b9b9bf]">
          <Bell className="size-[17px]" />
          <span className="absolute top-2 right-[9px] size-1.5 rounded-full border-[1.5px] border-[#121214] bg-[#f0484a]" />
        </button>
        {action && ActionIcon && (
          <button
            onClick={action.onClick}
            className="flex h-9 items-center gap-[7px] rounded-[10px] bg-[#f4f4f6] px-3.5 text-[13.5px] font-semibold text-[#0a0a0b] transition-colors hover:bg-white"
          >
            <ActionIcon className="size-[15px]" />
            {action.label}
          </button>
        )}
      </div>
    </div>
  );
}
