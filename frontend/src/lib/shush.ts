// Typed client for the Shush backend (ConnectRPC's HTTP/JSON surface).
//
// Calls are made server-side with the caller's Clerk session token in the
// Authorization header — the backend's auth interceptor verifies it. This is
// the BFF pattern: the browser never talks to the Go backend directly.

const API_URL = process.env.SHUSH_API_URL ?? "http://localhost:8080";

export type Project = {
  id: string;
  name: string;
  createdAt: string;
};

export type Secret = {
  key: string;
  // Connect omits proto3 zero-values from JSON, so an empty value / version 0
  // may be absent on the wire — treat these as optional on the client.
  value?: string;
  version?: number;
};

export const ENVIRONMENTS = ["dev", "staging", "prod"] as const;
export type Environment = (typeof ENVIRONMENTS)[number];

async function call<T>(method: string, token: string, body: unknown): Promise<T> {
  const res = await fetch(`${API_URL}/shush.v1.ShushService/${method}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(body ?? {}),
    cache: "no-store",
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`${method} failed (${res.status}): ${text}`);
  }
  return res.json() as Promise<T>;
}

export function listProjects(token: string) {
  return call<{ projects?: Project[] }>("ListProjects", token, {});
}

export function createProject(token: string, name: string) {
  return call<{ project: Project }>("CreateProject", token, { name });
}

export function listSecrets(
  token: string,
  projectId: string,
  environment: Environment,
) {
  return call<{ secrets?: Secret[] }>("ListSecrets", token, {
    projectId,
    environment,
  });
}

export function getSecret(
  token: string,
  projectId: string,
  environment: Environment,
  key: string,
) {
  return call<{ secret?: Secret }>("GetSecret", token, {
    projectId,
    environment,
    key,
  });
}

export function putSecret(
  token: string,
  projectId: string,
  environment: Environment,
  key: string,
  value: string,
) {
  return call<{ id: string; version: number }>("PutSecret", token, {
    projectId,
    environment,
    key,
    value,
  });
}
