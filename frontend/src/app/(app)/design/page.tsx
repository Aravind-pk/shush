import * as React from "react";
import { Plus, RefreshCw, Search } from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Progress } from "@/components/ui/progress";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { SeverityBadge, type Severity } from "@/components/severity";
import { Metric, SeverityStat } from "@/components/stat-card";

function Section({
  title,
  description,
  children,
}: {
  title: string;
  description?: string;
  children: React.ReactNode;
}) {
  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-lg font-semibold tracking-tight">{title}</h2>
        {description && (
          <p className="text-sm text-muted-foreground">{description}</p>
        )}
      </div>
      {children}
    </section>
  );
}

function Swatch({
  name,
  className,
  value,
}: {
  name: string;
  className: string;
  value?: string;
}) {
  return (
    <div className="space-y-1.5">
      <div className={cn("h-16 rounded-lg border border-border", className)} />
      <div className="text-xs font-medium">{name}</div>
      {value && (
        <div className="font-mono text-[11px] text-muted-foreground">
          {value}
        </div>
      )}
    </div>
  );
}

const SEVERITIES: Severity[] = ["critical", "high", "medium", "low"];

const CHART = [
  { name: "p50", cls: "bg-chart-1", hex: "#6366f1" },
  { name: "p75", cls: "bg-chart-2", hex: "#a855f7" },
  { name: "p95", cls: "bg-chart-3", hex: "#ec4899" },
  { name: "p99", cls: "bg-chart-4", hex: "#f97316" },
  { name: "max", cls: "bg-chart-5", hex: "#facc15" },
];

export default function DesignPage() {
  return (
    <main className="mx-auto w-full max-w-5xl space-y-12 px-6 py-10">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight">
          Shush Design System
        </h1>
        <p className="max-w-2xl text-sm text-muted-foreground">
          A dark, data-dense visual language inspired by error-tracking
          dashboards. Built on shadcn/ui + Tailwind v4 tokens. This page is a
          living reference — every color, component, and pattern below is wired
          to the same tokens the product uses.
        </p>
      </header>

      <Section
        title="Surfaces"
        description="Near-black ambient with subtly elevated cards and hairline borders."
      >
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6">
          <Swatch name="background" className="bg-background" />
          <Swatch name="card" className="bg-card" />
          <Swatch name="muted" className="bg-muted" />
          <Swatch name="accent" className="bg-accent" />
          <Swatch name="secondary" className="bg-secondary" />
          <Swatch name="primary" className="bg-primary" />
        </div>
      </Section>

      <Section
        title="Severity scale"
        description="The core semantic palette for issues."
      >
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          <Swatch name="critical" className="bg-critical" value="#f04349" />
          <Swatch name="high" className="bg-high" value="#f59e0b" />
          <Swatch name="medium" className="bg-medium" value="#eab308" />
          <Swatch name="low" className="bg-low" value="#3b82f6" />
        </div>
      </Section>

      <Section
        title="Chart palette"
        description="Cool→warm percentile gradient used across charts and timelines."
      >
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-5">
          {CHART.map((c) => (
            <Swatch key={c.name} name={c.name} className={c.cls} value={c.hex} />
          ))}
        </div>
      </Section>

      <Section title="Typography" description="Geist Sans for UI, Geist Mono for data.">
        <Card>
          <CardContent className="space-y-3 py-6">
            <p className="text-2xl font-semibold tracking-tight">
              Heading — 2xl semibold
            </p>
            <p className="text-lg font-semibold">Subheading — lg semibold</p>
            <p className="text-sm">Body — sm regular</p>
            <p className="text-sm text-muted-foreground">
              Muted label — sm muted-foreground
            </p>
            <p className="font-mono text-sm tabular-nums">
              Mono / numeric — 12.90s · #18369635b46c
            </p>
          </CardContent>
        </Card>
      </Section>

      <Section title="Buttons">
        <div className="flex flex-wrap items-center gap-3">
          <Button>Primary</Button>
          <Button variant="secondary">Secondary</Button>
          <Button variant="outline">Outline</Button>
          <Button variant="ghost">Ghost</Button>
          <Button variant="destructive">Destructive</Button>
          <Button size="icon" variant="outline" aria-label="Add">
            <Plus />
          </Button>
          <Button size="icon" variant="ghost" aria-label="Refresh">
            <RefreshCw />
          </Button>
        </div>
      </Section>

      <Section title="Badges">
        <div className="flex flex-wrap items-center gap-3">
          {SEVERITIES.map((s) => (
            <SeverityBadge key={s} severity={s} />
          ))}
          <Separator orientation="vertical" className="h-5" />
          <Badge>Default</Badge>
          <Badge variant="secondary">main-branch</Badge>
          <Badge variant="outline">v2.8.1</Badge>
        </div>
      </Section>

      <Section
        title="Severity stats"
        description="Header summary tiles."
      >
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
          <SeverityStat severity="critical" value={17} />
          <SeverityStat severity="high" value={83} />
          <SeverityStat severity="medium" value={64} />
          <SeverityStat severity="low" value={2} />
        </div>
      </Section>

      <Section title="Metrics & usage">
        <Card>
          <CardContent className="space-y-6 py-6">
            <div className="grid grid-cols-2 gap-6 sm:grid-cols-4">
              <Metric label="Minimum" value={16} unit="ms" />
              <Metric label="Maximum" value={82} unit="ms" />
              <Metric label="Average" value={137} unit="ms" />
              <Metric label="Median" value={29} unit="ms" />
            </div>
            <div className="space-y-2">
              <div className="flex items-center justify-between text-xs">
                <span className="text-medium">65% used</span>
                <span className="text-muted-foreground">35% free</span>
              </div>
              <Progress value={65} className="h-2 [&>div]:bg-medium" />
            </div>
          </CardContent>
        </Card>
      </Section>

      <Section title="Inputs & tabs">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
          <div className="relative w-full max-w-xs">
            <Search className="absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input placeholder="Search spans…" className="pl-8" />
          </div>
          <Tabs defaultValue="waterfall">
            <TabsList>
              <TabsTrigger value="waterfall">Waterfall</TabsTrigger>
              <TabsTrigger value="attributes">Attributes</TabsTrigger>
              <TabsTrigger value="profiles">Profiles</TabsTrigger>
            </TabsList>
          </Tabs>
        </div>
      </Section>

      <Section title="Table">
        <Card className="overflow-hidden py-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Issue</TableHead>
                <TableHead>Severity</TableHead>
                <TableHead className="text-right">Events</TableHead>
                <TableHead>Assignee</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {[
                ["NullPointerException", "critical", 234],
                ["Timeout on /checkout", "high", 88],
                ["Slow DB query", "medium", 51],
                ["Deprecated API usage", "low", 3],
              ].map(([issue, sev, events]) => (
                <TableRow key={issue as string}>
                  <TableCell className="font-mono text-xs">{issue}</TableCell>
                  <TableCell>
                    <SeverityBadge severity={sev as Severity} />
                  </TableCell>
                  <TableCell className="text-right font-mono tabular-nums">
                    {events}
                  </TableCell>
                  <TableCell>
                    <Avatar className="size-6">
                      <AvatarFallback className="text-[10px]">AP</AvatarFallback>
                    </Avatar>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Card>
      </Section>

      <Section title="Card & tooltip">
        <Card>
          <CardHeader>
            <CardTitle>Error Duration Percentiles</CardTitle>
            <CardDescription>
              Hover the metric for an explanation.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="outline" size="sm">
                  What is p95?
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                95% of requests completed faster than this value.
              </TooltipContent>
            </Tooltip>
          </CardContent>
        </Card>
      </Section>
    </main>
  );
}
