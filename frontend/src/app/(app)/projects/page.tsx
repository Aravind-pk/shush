import Link from "next/link";
import { auth } from "@clerk/nextjs/server";
import { ArrowUpRight, FolderClosed, KeyRound, Search } from "lucide-react";

import { listProjects, type Project } from "@/lib/shush";
import { ENV_META } from "@/lib/env-meta";
import { CreateProjectDialog } from "@/components/create-project-dialog";
import { ProjectsChrome } from "@/components/projects-chrome";

export default async function ProjectsPage() {
  const { getToken } = await auth();
  const token = await getToken();

  let projects: Project[] = [];
  let error: string | null = null;

  if (token) {
    try {
      const res = await listProjects(token);
      projects = res.projects ?? [];
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
    }
  }

  return (
    <div className="flex flex-col gap-[22px]">
      {/* Populates the top bar breadcrumb + "New project" action */}
      <ProjectsChrome />

      {/* Header */}
      <div className="flex items-center justify-between border-b border-white/[0.07] pb-5">
        <div>
          <h1 className="text-[23px] font-bold tracking-[-0.02em] text-[#fafafa]">
            Projects
          </h1>
          <p className="mt-1 text-sm text-[#9a9aa0]">
            {projects.length} project{projects.length === 1 ? "" : "s"} · 3
            environments each
          </p>
        </div>
        <div className="flex h-[38px] w-[260px] items-center gap-[9px] rounded-[10px] border border-white/[0.09] bg-[#121214] px-3">
          <Search className="size-[15px] text-[#6a6a70]" />
          <span className="text-[13px] text-[#6a6a70]">Search projects…</span>
        </div>
      </div>

      {error && (
        <div className="rounded-[14px] border border-critical/30 bg-critical/10 p-4 text-sm text-critical">
          Backend error: {error}
        </div>
      )}

      {!error && projects.length === 0 && (
        <div className="flex flex-col items-center gap-3 rounded-[14px] border border-white/[0.07] bg-[linear-gradient(180deg,#131315,#0f0f11)] px-6 py-16 text-center">
          <span className="flex size-12 items-center justify-center rounded-xl bg-white/[0.04]">
            <FolderClosed className="size-6 text-[#7d7d84]" />
          </span>
          <p className="text-[15px] font-medium text-[#e9e9ec]">
            No projects yet
          </p>
          <p className="max-w-sm text-sm text-[#7d7d84]">
            Create your first project to start storing encrypted secrets per
            environment.
          </p>
          <div className="mt-2">
            <CreateProjectDialog />
          </div>
        </div>
      )}

      {projects.length > 0 && (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {projects.map((p) => (
            <Link
              key={p.id}
              href={`/projects/${p.id}`}
              className="group rounded-[14px] border border-white/[0.07] bg-[linear-gradient(180deg,#131315,#0f0f11)] p-5 transition-colors hover:border-white/[0.16]"
            >
              <div className="mb-[14px] flex items-center justify-between">
                <div className="flex items-center gap-[11px]">
                  <span className="flex size-[38px] items-center justify-center rounded-[10px] bg-[rgba(123,84,196,0.13)]">
                    <KeyRound className="size-[19px] text-[#b08ef0]" />
                  </span>
                  <div>
                    <div className="font-mono text-[15px] font-semibold text-[#f4f4f6]">
                      {p.name}
                    </div>
                    <div className="text-xs text-[#7d7d84]">
                      Created {formatDate(p.createdAt)}
                    </div>
                  </div>
                </div>
                <ArrowUpRight className="size-4 text-[#5f5f66] transition-colors group-hover:text-[#cfcfd5]" />
              </div>
              <div className="flex gap-1.5">
                {(Object.keys(ENV_META) as (keyof typeof ENV_META)[]).map(
                  (env) => (
                    <span
                      key={env}
                      className="flex items-center gap-[5px] rounded-[6px] border border-white/[0.06] bg-[#1a1a1d] px-2 py-[3px] font-mono text-[11.5px] text-[#9a9aa0]"
                    >
                      <span
                        className="size-1.5 rounded-full"
                        style={{ background: ENV_META[env].dot }}
                      />
                      {ENV_META[env].label}
                    </span>
                  ),
                )}
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}

function formatDate(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}
