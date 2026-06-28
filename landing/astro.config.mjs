// @ts-check
import { defineConfig } from "astro/config";
import tailwindcss from "@tailwindcss/vite";
import icon from "astro-icon";

// Static marketing site for Shush. No auth, no backend — builds to static HTML.
//
// Deployed to GitHub Pages as a *project page* at
//   https://aravind-pk.github.io/shush/
// hence `base: "/shush"`. If you move to a custom domain (e.g. shush.dev),
// set `site` to that domain and change `base` back to "/".
export default defineConfig({
  site: "https://aravind-pk.github.io",
  base: "/shush",
  vite: {
    plugins: [tailwindcss()],
  },
  integrations: [icon()],
});
