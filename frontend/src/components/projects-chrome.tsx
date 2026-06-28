"use client";

import * as React from "react";
import { Plus } from "lucide-react";

import { useSetChrome, type Crumb, type ChromeAction } from "@/components/app-chrome";
import { CreateProjectDialog } from "@/components/create-project-dialog";

const BREADCRUMB: Crumb[] = [{ label: "Shush" }, { label: "Projects" }];

/**
 * Wires the Projects screen into the top bar: a "Projects" breadcrumb and a
 * "New project" primary action that opens the create dialog.
 */
export function ProjectsChrome() {
  const [open, setOpen] = React.useState(false);

  const action = React.useMemo<ChromeAction>(
    () => ({ label: "New project", icon: Plus, onClick: () => setOpen(true) }),
    [],
  );
  useSetChrome(BREADCRUMB, action);

  return (
    <CreateProjectDialog open={open} onOpenChange={setOpen} trigger={null} />
  );
}
