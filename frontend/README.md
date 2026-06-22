# Dozingo Frontend

Vue 3 SPA for the Dozingo bingo game.

---

## Tech Stack

| Layer                | Technology                            |
| -------------------- | ------------------------------------- |
| Framework            | Vue 3 (Composition API)               |
| Language             | TypeScript                            |
| Build Tool           | Vite                                  |
| Routing              | vue-router v5                         |
| Localisation         | vue-i18n v11 (English / German)       |
| Icons                | lucide-vue-next                       |
| Serving (production) | Caddy (see `Dockerfile`, `Caddyfile`) |

## Development

```bash
npm install
npm run dev
```

The dev server runs at [http://localhost:5173](http://localhost:5173) and
proxies `/api` requests to the backend at `http://localhost:4242`.

```bash
npm run build      # type-check + production build
npm run preview    # serve the production build locally
```

## Production Image

The `Dockerfile` in this directory produces a self-contained image:

- **Build stage**: Node 22 runs `npm ci && npm run build`
- **Runtime stage**: Caddy serves the `dist/` assets on port 4242, falling back
  to `index.html` for client-side routing, and reverse-proxies `/api/*` to the
  backend service

Built and pushed to GHCR automatically on push to `main` via
`.github/workflows/frontend-image.yml`.
