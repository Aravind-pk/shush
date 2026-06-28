"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { UserButton } from "@clerk/nextjs";
import {
  Asterisk,
  ChevronsUpDown,
  LayoutDashboard,
  Folder,
  Users,
  UserRound,
  ScrollText,
  Settings,
  type LucideIcon,
} from "lucide-react";

import { cn } from "@/lib/utils";

type NavItem = {
  label: string;
  href: string;
  icon: LucideIcon;
  /** Routes not built yet are shown disabled. */
  soon?: boolean;
};

const NAV: NavItem[] = [
  { label: "Overview", href: "/dashboard", icon: LayoutDashboard, soon: true },
  { label: "Projects", href: "/projects", icon: Folder },
  { label: "Groups", href: "/groups", icon: Users, soon: true },
  { label: "Members", href: "/members", icon: UserRound, soon: true },
  { label: "Audit Log", href: "/audit", icon: ScrollText, soon: true },
];

export function AppSidebar() {
  const pathname = usePathname();

  return (
    <aside className="sticky top-0 flex h-screen w-[230px] flex-none flex-col border-r border-white/[0.06] bg-[#08080a] px-[14px] py-[18px]">
      {/* Workspace / org switcher (static until the org model lands) */}
      <button className="mb-[18px] flex w-full items-center gap-[11px] rounded-[11px] border border-white/[0.08] bg-[#121214] px-[11px] py-[9px] text-left transition-colors hover:border-white/[0.18]">
        <span className="flex size-8 flex-none items-center justify-center rounded-[9px] bg-[radial-gradient(circle_at_35%_30%,#f5c518,#f5841f_45%,#a64ba6_80%)]">
          <Asterisk className="size-[17px] text-[#0a0a0b]" strokeWidth={2.5} />
        </span>
        <span className="min-w-0 flex-1">
          <span className="block text-sm font-semibold text-[#fafafa]">
            Shush
          </span>
          <span className="block font-mono text-[11.5px] text-[#7d7d84]">
            workspace
          </span>
        </span>
        <ChevronsUpDown className="size-[15px] flex-none text-[#7d7d84]" />
      </button>

      <nav className="flex flex-col gap-[3px]">
        {NAV.map((item) => {
          const active =
            pathname === item.href || pathname.startsWith(item.href + "/");
          const Icon = item.icon;

          if (item.soon) {
            return (
              <span
                key={item.href}
                title="Coming soon"
                className="flex w-full cursor-not-allowed items-center gap-[11px] rounded-[9px] px-[11px] py-[9px] text-sm font-medium text-[#5f5f66]"
              >
                <Icon className="size-[18px] flex-none" />
                <span className="flex-1">{item.label}</span>
                <span className="rounded-[5px] bg-white/[0.04] px-1.5 py-px font-mono text-[10px] text-[#6a6a70]">
                  soon
                </span>
              </span>
            );
          }

          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex w-full items-center gap-[11px] rounded-[9px] px-[11px] py-[9px] text-sm font-medium transition-colors",
                active
                  ? "bg-white/[0.08] text-[#fafafa]"
                  : "text-[#8a8a90] hover:bg-white/[0.05]",
              )}
            >
              <Icon className="size-[18px] flex-none" />
              <span className="flex-1">{item.label}</span>
            </Link>
          );
        })}
      </nav>

      <div className="flex-1" />

      <div className="flex flex-col gap-[3px] border-t border-white/[0.06] pt-[10px]">
        <span className="flex w-full cursor-not-allowed items-center gap-[11px] rounded-[9px] px-[11px] py-[9px] text-sm text-[#5f5f66]">
          <Settings className="size-[18px]" />
          Settings
        </span>
        <div className="flex items-center gap-[10px] px-[11px] py-[9px]">
          <UserButton
            showName
            appearance={{
              elements: {
                rootBox: "w-full",
                userButtonBox: "flex-row-reverse w-full justify-end gap-2.5",
                userButtonOuterIdentifier: "text-[13px] text-[#dcdce0]",
              },
            }}
          />
        </div>
      </div>
    </aside>
  );
}
