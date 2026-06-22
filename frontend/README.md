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
