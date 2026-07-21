# TorrServer-Go — Web session context

**Saved:** 2026-07-18  
**Branch (typical):** `feature/web-upgrade`  
**Phase:** `web/` full-scratch rebuild complete — all features on HeroUI v3 + GSAP; typecheck/lint/test/build/gen_web green.

---

## Doctrine

- `web/` is a from-scratch client: `app/` + `features/` + `shared/` + `locales/`. Every UI file was rewritten fresh, not ported.
- New client to TorrServer API — not visual parity with any prior UI
- Server/API = feature/behavior contract only
- Prefer `shared/api/*` + React Query

---

## Stack

| Layer | Value |
|-------|--------|
| Runtime | React 19 + Vite 8.1 + TypeScript 6 + Vitest 4 |
| UI | HeroUI 3 (`@heroui/react` + `@heroui/styles`) + Tailwind 4 |
| Motion | GSAP + `@gsap/react` |
| Icons / toast | lucide-react / sonner |
| Theme | `useThemePreference` — class on `html` |
| Ship | typecheck + lint + test + build → `gen_web --clean` |

---

## Done

- `web/` scaffolded from scratch (package.json, vite/tsconfig/eslint, index.html, public/ assets) — fresh `yarn.lock`.
- Functional layer in place: `shared/api`, `shared/hooks`, `shared/lib`, `shared/torrent`, `shared/cache`, `shared/i18n` + `locales` (en/ru/ua/zh/bg/fr/ro), `shared/theme/{breakpoints,color,useThemePreference}`, `shared/settings`, `shared/ui` (AppDialog/ModalOpenContext/Toast/UnsafeButton).
- `index.css` brand theme (HeroUI CSS-var overrides, light+dark); `app/{App,Shell,Sidebar,BottomNav,ErrorBoundary}` rewritten.
- Features: `torrents` (poster grid + GSAP), `player` (HLS/gstreamer VideoPlayer), `details` (overview/files/cache tabs, snake canvas), `add` + `search`, `settings` (+ Torznab/TMDB) + `about`/`system`/`categories`/`pwa`.
- Full pipeline green: `yarn typecheck && yarn lint && yarn test && yarn build`, `go run gen_web.go --clean` into `server/web/pages/template/pages`, Go server builds clean.

---

## Agent rules

- Russian replies when user writes Russian; code English.
- No commit unless asked.
