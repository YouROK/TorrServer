---
name: torrserver-web
description: >-
  Continues TorrServer-Go web UI work (React 19, Vite 8, TypeScript 6, HeroUI 3,
  GSAP, TanStack Query, gen_web embed). Use when editing web/, torrent
  details/cache, GStreamer player, empty states, i18n, or web upgrade.
---

# TorrServer-Go Web

## Before coding

1. Read [web-upgrade-session.md](../../context/web-upgrade-session.md) for locked decisions.
2. Confirm Node 22+ (`web/.nvmrc`). Use **yarn** in `web/`.

## Hard locks

- Keep React — no Vue rewrite.
- **Tooling:** Vite **8.1** + TypeScript **6.0** + Vitest **4** + `@vitejs/plugin-react` **6**. Node **22+**.
- **UI:** **HeroUI v3** (`@heroui/react` + `@heroui/styles`) + **Tailwind CSS v4** (HeroUI peer) + **GSAP** (`gsap` + `@gsap/react`) + **lucide-react** + **sonner**.
- **No MUI / Emotion / styled-components.**
- **Greenfield tree:** `web/src/{app,features,shared,locales}` only.
- **Doctrine:** new web client to TorrServer API — rewrite UI freely; server/API = feature/behavior contract only; never port old visuals.
- Prefer `shared/api/*` + React Query hooks.
- Brand accent OK; streaming poster library; Safari **17+**.
- Donate restored (dialog + snackbar + nav). Adaptive shell + `ModalOpenProvider`.
- Web Basic login (AuthGate + LoginScreen); SPA public; third-party API/stream keep sending Authorization.
- Snake/GStreamer contracts unchanged. File row: primary actions reachable.

## Layout

```
web/src/app/            # App, Shell, Sidebar, BottomNav, ErrorBoundary
web/src/features/       # torrents, add, search, settings, details, player, about, system, pwa, categories
web/src/shared/         # api, hooks, theme, cache, lib, ui, i18n, torrent, settings
web/src/locales/        # i18n JSON
```

## Workflow

```bash
cd web && yarn typecheck && yarn lint && yarn test && yarn build
cd .. && go run gen_web.go --clean
```

Always use `--clean`. Restart TorrServer + hard refresh. Do not commit unless asked.

## Patterns

| Need | Prefer |
|------|--------|
| Toast | `shared/ui/Toast` (sonner) |
| Dialog | `shared/ui/AppDialog` (HeroUI Modal + useOverlayState) |
| Torrents list | `features/torrents` poster grid + GSAP |
| Theme | `useThemePreference` — `dark`/`light` class on `html` |
| Modal open / immersive | `useSyncModalOpen` / `setImmersive` |

## Prompt for new chats

Paste from [PROMPT.md](PROMPT.md).
