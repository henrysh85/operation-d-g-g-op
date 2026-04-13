# DCGG Intelligence Platform — Frontend

Vue 3 + TypeScript + Vite SPA for the DCGG Intelligence Platform.

## Stack
- Vue 3 `<script setup lang="ts">`
- Vite, Vue Router, Pinia, Axios
- TailwindCSS, Headless UI, date-fns, @vueuse/core

## Requirements
- Node 20+

## Getting started

```bash
npm install
npm run dev        # Vite dev server on :5173, /api proxied to :8080
npm run typecheck
npm run lint
npm run build      # outputs /dist
npm run preview
```

## Project layout

```
src/
  api/         Axios resource modules (one per backend entity)
  assets/      tokens.css — design tokens from prototype
  components/  Sidebar, FilterBar, DataTable, badges, OrgChart, Modal
  router/      All 11 modules + /login + /members/:id
  stores/      Pinia: auth, filters
  types/       Backend model interfaces
  views/       One per route
```

## Environment

`/api` is proxied to `http://localhost:8080` in dev (see `vite.config.ts`).
In production, nginx proxies `/api/` to `http://backend:8080/api/`
(see `nginx.conf`).

Override with `VITE_API_BASE` at build time if needed.

## Docker

```bash
docker build -t dcgg-frontend .
docker run -p 8081:80 dcgg-frontend
```
