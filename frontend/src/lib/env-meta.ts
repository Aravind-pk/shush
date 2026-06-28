import type { Environment } from "@/lib/shush";

// Per-environment display metadata: label + the coloured dot used in the env
// switcher and chips throughout the dashboard.
export const ENV_META: Record<
  Environment,
  { label: string; dot: string; text: string }
> = {
  dev: { label: "dev", dot: "#3ddc84", text: "#3ddc84" },
  staging: { label: "staging", dot: "#f6912b", text: "#f6912b" },
  prod: { label: "prod", dot: "#ef5d83", text: "#ef5d83" },
};
