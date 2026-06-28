"use server";

import { auth } from "@clerk/nextjs/server";
import { revalidatePath } from "next/cache";
import {
  createProject,
  getSecret,
  putSecret,
  type Environment,
} from "@/lib/shush";

// Server Actions are reachable via direct POST, so each one re-checks auth and
// resolves the Clerk session token before calling the backend RPC.
async function token() {
  const { getToken } = await auth();
  const t = await getToken();
  if (!t) throw new Error("Not authenticated");
  return t;
}

export type ActionResult = { ok: true } | { ok: false; error: string };

export async function createProjectAction(name: string): Promise<ActionResult> {
  const trimmed = name.trim();
  if (!trimmed) return { ok: false, error: "Project name is required" };
  try {
    await createProject(await token(), trimmed);
    revalidatePath("/projects");
    return { ok: true };
  } catch (e) {
    return { ok: false, error: e instanceof Error ? e.message : String(e) };
  }
}

export type RevealResult =
  | { ok: true; value: string }
  | { ok: false; error: string };

// Fetches a single secret's plaintext on demand (the "reveal" / copy path).
// Each call hits the backend's GetSecret RPC — the only path that decrypts —
// so values are disclosed one at a time rather than on list.
export async function getSecretAction(
  projectId: string,
  environment: Environment,
  key: string,
): Promise<RevealResult> {
  try {
    const res = await getSecret(await token(), projectId, environment, key);
    return { ok: true, value: res.secret?.value ?? "" };
  } catch (e) {
    return { ok: false, error: e instanceof Error ? e.message : String(e) };
  }
}

export async function putSecretAction(
  projectId: string,
  environment: Environment,
  key: string,
  value: string,
): Promise<ActionResult> {
  const trimmedKey = key.trim();
  if (!trimmedKey) return { ok: false, error: "Key is required" };
  try {
    await putSecret(await token(), projectId, environment, trimmedKey, value);
    revalidatePath(`/projects/${projectId}`);
    return { ok: true };
  } catch (e) {
    return { ok: false, error: e instanceof Error ? e.message : String(e) };
  }
}
