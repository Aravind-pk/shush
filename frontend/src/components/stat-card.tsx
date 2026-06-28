import * as React from "react";

import { cn } from "@/lib/utils";
import { Card } from "@/components/ui/card";
import { SEVERITY, type Severity } from "@/components/severity";

/**
 * Severity summary tile used in the dashboard header row:
 * an icon + label on the left, a large count on the right.
 */
export function SeverityStat({
  severity,
  value,
  className,
}: {
  severity: Severity;
  value: number | string;
  className?: string;
}) {
  const meta = SEVERITY[severity];
  const Icon = meta.icon;
  return (
    <Card
      className={cn(
        "flex flex-row items-center justify-between gap-4 px-5 py-4",
        className,
      )}
    >
      <span className="flex items-center gap-2 text-sm text-muted-foreground">
        <Icon className={cn("size-4", meta.text)} />
        {meta.label}
      </span>
      <span className="font-mono text-2xl font-semibold tabular-nums text-foreground">
        {value}
      </span>
    </Card>
  );
}

/** Generic labelled metric (e.g. Average / Median / Root Duration). */
export function Metric({
  label,
  value,
  unit,
  className,
}: {
  label: string;
  value: string | number;
  unit?: string;
  className?: string;
}) {
  return (
    <div className={cn("flex flex-col gap-1", className)}>
      <span className="text-xs text-muted-foreground">{label}</span>
      <span className="font-mono text-xl font-semibold tabular-nums text-foreground">
        {value}
        {unit && (
          <span className="ml-1 text-sm font-normal text-muted-foreground">
            {unit}
          </span>
        )}
      </span>
    </div>
  );
}
