import Link from "next/link";
import { notFound } from "next/navigation";
import { auth } from "@clerk/nextjs/server";
import { Download, Folder, Lock, Upload } from "lucide-react";

import {
  listProjects,
  listSecrets,
  ENVIRONMENTS,
  type Environment,
  type Secret,
} from "@/lib/shush";
import { ENV_META } from "@/lib/env-meta";
import { SecretsChrome } from "@/components/secrets-chrome";
import { SecretsTable } from "@/components/secrets-table";

function isEnvironment(v: string | undefined): v is Environment {
  return !!v && (ENVIRONMENTS as readonly string[]).includes(v);
}

export default async function ProjectSecretsPage({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>;
  searchParams: Promise<{ env?: string }>;
}) {
  const { id } = await params;
  const { env: envParam } = await searchParams;
  const environment: Environment = isEnvironment(envParam) ? envParam : "prod";

  const { getToken } = await auth();
  const token = await getToken();
  if (!token) notFound();

  // No GetProject RPC yet — resolve the project name from the user's list.
  const projectsRes = await listProjects(token);
  const project = (projectsRes.projects ?? []).find((p) => p.id === id);
  if (!project) notFound();

  let secrets: Secret[] = [];
  let error: string | null = null;
  try {
    const res = await listSecrets(token, id, environment);
    secrets = res.secrets ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : String(e);
  }

  return (
    <div className="flex flex-col gap-5">
      {/* Populates the top bar breadcrumb + "Add secret" action */}
      <SecretsChrome
        projectId={id}
        projectName={project.name}
        environment={environment}
      />

      {/* Header */}
      <div className="flex items-center justify-between border-b border-white/[0.07] pb-[18px]">
        <div className="flex items-center gap-[14px]">
          <span className="flex size-[42px] items-center justify-center rounded-[11px] bg-[rgba(123,84,196,0.13)]">
            <Folder className="size-[21px] text-[#b08ef0]" />
          </span>
          <div>
            <div className="font-mono text-[23px] font-bold tracking-[-0.02em] text-[#fafafa]">
              {project.name}
            </div>
            <div className="mt-0.5 text-[13.5px] text-[#9a9aa0]">
              Encrypted secrets, scoped per environment
            </div>
          </div>
        </div>
        <div className="flex gap-[10px]">
          <button
            disabled
            title="Coming soon"
            className="flex h-9 cursor-not-allowed items-center gap-[7px] rounded-[10px] border border-white/[0.09] bg-[#121214] px-[13px] text-[13.5px] font-medium text-[#cfcfd5] opacity-60"
          >
            <Upload className="size-[15px]" />
            Import .env
          </button>
          <button
            disabled
            title="Coming soon"
            className="flex h-9 cursor-not-allowed items-center gap-[7px] rounded-[10px] border border-white/[0.09] bg-[#121214] px-[13px] text-[13.5px] font-medium text-[#cfcfd5] opacity-60"
          >
            <Download className="size-[15px]" />
            Export
          </button>
        </div>
      </div>

      {/* Env switcher + status line */}
      <div className="flex items-center justify-between">
        <div className="flex gap-1.5 rounded-[11px] border border-white/[0.08] bg-[#121214] p-1">
          {ENVIRONMENTS.map((env) => {
            const active = env === environment;
            return (
              <Link
                key={env}
                href={`/projects/${id}?env=${env}`}
                className={
                  "flex items-center gap-[7px] rounded-lg px-3.5 py-[7px] font-mono text-[13px] font-medium transition-colors " +
                  (active
                    ? "bg-white/[0.1] text-[#fafafa]"
                    : "text-[#8a8a90] hover:bg-white/[0.04]")
                }
              >
                <span
                  className="size-[7px] rounded-full"
                  style={{ background: ENV_META[env].dot }}
                />
                {ENV_META[env].label}
              </Link>
            );
          })}
        </div>
        <div className="flex items-center gap-2 text-[13px] text-[#7d7d84]">
          <Lock className="size-[15px] text-[#3ddc84]" />
          AES-256-GCM · {secrets.length} secret
          {secrets.length === 1 ? "" : "s"} in {environment}
        </div>
      </div>

      {error ? (
        <div className="rounded-[14px] border border-critical/30 bg-critical/10 p-4 text-sm text-critical">
          Backend error: {error}
        </div>
      ) : (
        <SecretsTable
          secrets={secrets}
          projectId={id}
          environment={environment}
        />
      )}
    </div>
  );
}
