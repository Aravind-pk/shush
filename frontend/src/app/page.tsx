import { redirect } from "next/navigation";

// The marketing landing page lives in the standalone `landing/` site.
// The Next app is the authenticated dashboard — send the root to it.
export default function Home() {
  redirect("/projects");
}
