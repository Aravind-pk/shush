# Shush Landing

The public marketing site for Shush. **Standalone and static** — it has no auth,
no backend calls, and is intentionally decoupled from the `frontend/` dashboard
app (which is Clerk-gated). Deploy it anywhere that serves static files
(GitHub Pages, Cloudflare Pages, a CDN bucket).

## Stack

- [Astro](https://astro.build) — static output, ~zero client JS
- [Tailwind CSS v4](https://tailwindcss.com) via `@tailwindcss/vite`
- [astro-icon](https://github.com/natemoo-re/astro-icon) + `@iconify-json/lucide`

## Develop

```bash
npm install
npm run dev      # http://localhost:4321
npm run build    # static output to dist/
npm run preview  # serve the built site
```

## Structure

- `src/styles/global.css` — brand design tokens (colors, fonts) + component
  utilities (`.card`, `.term`, `.btn`, …). Edit tokens here, not in components.
- `src/layouts/Base.astro` — `<head>`, fonts, global CSS.
- `src/components/` — one component per page section (`Hero`, `Problem`,
  `BeforeAfter`, `HowItWorks`, `Onboarding`, `Features`, `Principles`, `CTA`),
  plus reusable `Nav`, `Footer`, `Logo`, `Terminal`, `SectionHeading`.
- `src/pages/index.astro` — composes the sections.

## Notes

- Design reference came from the Claude Design project "Shush Landing"; this is a
  componentized reimplementation, not a verbatim export.
- Outbound links (`GITHUB`, `DOCS`) are defined at the top of `Nav`, `Hero`,
  `CTA`, and `Footer` — update them when those destinations are real.
