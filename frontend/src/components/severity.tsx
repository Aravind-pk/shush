import * as React from "react";
import {
  OctagonAlert,
  TriangleAlert,
  CircleAlert,
  Info,
  type LucideIcon,
} from "lucide-react";

import { cn } from "@/lib/utils";

export type Severity = "critical" | "high" | "medium" | "low";

type SeverityMeta = {
  label: string;
  icon: LucideIcon;
  /** text color utility */
  text: string;
  /** soft tinted badge utility */
  badge: string;
  /** solid dot utility */
  dot: string;
};

export const SEVERITY: Record<Severity, SeverityMeta> = {
  critical: {
    label: "Critical",
    icon: OctagonAlert,
    text: "text-critical",
    badge: "bg-critical/10 text-critical border-critical/20",
    dot: "bg-critical",
  },
  high: {
    label: "High",
    icon: TriangleAlert,
    text: "text-high",
    badge: "bg-high/10 text-high border-high/20",
    dot: "bg-high",
  },
  medium: {
    label: "Medium",
    icon: CircleAlert,
    text: "text-medium",
    badge: "bg-medium/10 text-medium border-medium/20",
    dot: "bg-medium",
  },
  low: {
    label: "Low",
    icon: Info,
    text: "text-low",
    badge: "bg-low/10 text-low border-low/20",
    dot: "bg-low",
  },
};

export function SeverityBadge({
  severity,
  className,
  showIcon = true,
}: {
  severity: Severity;
  className?: string;
  showIcon?: boolean;
}) {
  const meta = SEVERITY[severity];
  const Icon = meta.icon;
  return (
    <span
      className={cn(
        "inline-flex h-5 w-fit shrink-0 items-center gap-1 rounded-4xl border px-2 py-0.5 text-xs font-medium",
        meta.badge,
        className,
      )}
    >
      {showIcon && <Icon className="size-3" />}
      {meta.label}
    </span>
  );
}
