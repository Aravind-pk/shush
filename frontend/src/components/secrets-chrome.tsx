"use client";

import * as React from "react";
import { Plus } from "lucide-react";

import { useSetChrome, type Crumb, type ChromeAction } from "@/components/app-chrome";
import { AddSecretDialog } from "@/components/add-secret-dialog";
import type { Environment } from "@/lib/shush";

/**
 * Wires the Secrets screen into the top bar: a `Shush / project / Secrets`
 * breadcrumb and an "Add secret" primary action that opens the add dialog.
 */
export function SecretsChrome({
  projectId,
  projectName,
  environment,
}: {
  projectId: string;
  projectName: string;
  environment: Environment;
}) {
  const [open, setOpen] = React.useState(false);

  const breadcrumb = React.useMemo<Crumb[]>(
    () => [{ label: "Shush" }, { label: projectName, mono: true }, { label: "Secrets" }],
    [projectName],
  );
  const action = React.useMemo<ChromeAction>(
    () => ({ label: "Add secret", icon: Plus, onClick: () => setOpen(true) }),
    [],
  );
  useSetChrome(breadcrumb, action);

  return (
    <AddSecretDialog
      projectId={projectId}
      environment={environment}
      open={open}
      onOpenChange={setOpen}
      trigger={null}
    />
  );
}
