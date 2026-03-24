# Architecture

## Stack
- **Frontend:** Svelte 5 + SvelteKit + Threlte (Three.js) + TypeScript
- **Backend:** Go + Gorilla WebSocket + SQLite
- **DevOps:** Docker + Colima locally, GitHub Actions → GHCR → Coolify in prod

## Game State Machine

```
WAITING → place piece → boop check → graduation check
  ├── no lines, <8 placed → WAITING (turnNumber++)
  ├── 1 line → auto-graduate → WAITING (turnNumber++)
  ├── >1 lines → MULTIPLE_WAITING (turnNumber unchanged)
  │     └── player selects line → graduate → WAITING (turnNumber++)
  └── no lines, 8 placed → MAX_WAITING (turnNumber unchanged)
        └── player selects piece → graduate → WAITING (turnNumber++)
```

`turnNumber` only increments when state returns to WAITING. `broadcastSeq` increments on every broadcast and is used for frontend deduplication.

After graduation handlers (MULTIPLE_WAITING/MAX_WAITING → WAITING), the server clears `Placed`, `Lines`, `BoopMovement`, and `Booped` to prevent stale data from re-triggering animations.

## Broadcast Data

Each broadcast from the server contains:

| Field | Description |
|---|---|
| `broadcastSeq` | Monotonic counter, increments every broadcast (frontend dedup) |
| `turnNumber` | Game turn counter (only increments on WAITING transitions) |
| `board` | Current 6x6 grid |
| `previousBoard` | Board state before this turn |
| `placed` | {position, piece} of piece just placed (cleared after graduation) |
| `boopMovement` | Pieces that slid on-board (cleared after graduation) |
| `booped` | Pieces knocked off-board (cleared after graduation) |
| `lines` | Detected 3-in-a-rows (cleared after graduation) |
| `threeChoices` | Middle positions for MULTIPLE_WAITING selection |

## Frontend Animation Trigger Logic

The message handler in `GameBrowser.svelte` determines which animations to trigger based on the **old state → new state transition**, not raw payload data.

### State Transition → Animation Map

| Old State | New State | Animations |
|---|---|---|
| WAITING | WAITING | Placement arc + slides + flying + auto-graduation |
| WAITING | MULTIPLE_WAITING | Placement arc + slides + flying |
| WAITING | MAX_WAITING | Placement arc + slides + flying |
| MULTIPLE_WAITING | WAITING | Graduation animation (board diff finds 3 removed pieces) |
| MAX_WAITING | WAITING | Single-piece graduation (board diff, same position ×3) |
| Any | Same broadcastSeq | Skip (dedup) |

### Dedup

In local pass-and-play, both `p1WebSocket` and `p2WebSocket` receive every broadcast. `broadcastSeq` prevents double-processing. The old approach used `turnNumber` which broke for MULTIPLE_WAITING/MAX_WAITING transitions where the turn doesn't increment.

### Arc Trigger

Kitten/Cat components react to the `$arcTrigger` store, **not** `$gameState.placed`. The `$arcTrigger` store is only set during real placements (WAITING → X transitions). This prevents graduation selections from re-triggering placement arcs, since `$gameState.placed` persists across broadcasts and would cause reactive `$effect`s to fire on every state update.

## Animation System

All 3D animations use Threlte's `useTask` (frame-by-frame), not the motion library (`animate` from motion breaks with Reduced Motion).

| Event | Component | Duration |
|---|---|---|
| Piece placed | Kitten/Cat.svelte | configurable (default 0.6s) |
| On-board boop | SlidingPiece.svelte | configurable (default 0.25s) |
| Off-board boop | FlyingPiece.svelte | ~1.5s |
| 3-piece graduation | GraduatingLine.svelte | 1.45s |
| Single-piece graduation | GraduatingLine.svelte (positions identical) | 1.45s |

### Sequencing

All timing is configurable via `$animConfig` store and the AnimDebug panel (gear icon, bottom-right):

1. Placement arc fires immediately
2. `$placementLanded` signals at `arcLandThreshold`% of the arc
3. Slides start after `$placementLanded` + `slideDelay` (negative = overlap)
4. Flying starts after `$placementLanded` + `flyDelay` (negative = overlap)

## Key Files

| File | Purpose |
|---|---|
| `logic/logic.go` | Game engine: boop/graduation logic, state machine |
| `logic/main.go` | WebSocket handlers, processTurn, readPump/writePump, broadcastGameState |
| `src/lib/components/Board.svelte` | 3D board, piece rendering, click handling |
| `src/lib/components/GameBrowser.svelte` | Lobby + animation trigger logic (state transition handler) |
| `src/lib/components/stores.ts` | Centralized Svelte stores (`arcTrigger`, `animConfig`, game state) |
| `src/lib/components/AnimDebug.svelte` | Animation tuning debug panel |
| `src/lib/components/{Kitten,Cat}.svelte` | Piece components with arc animation |
| `src/lib/components/{SlidingPiece,FlyingPiece,GraduatingLine}.svelte` | Animation components |

## Deployment Pipeline

### Flow

```
git push main
  → GitHub Actions builds frontend + backend Docker images
  → Pushes to GHCR with :main and :<commit-sha> tags
  → Updates Coolify TAG env var to commit SHA (delete + create via API)
  → Triggers Coolify deploy webhook
  → Coolify runs docker-compose with TAG=<sha>, pulling fresh images
```

### Coolify Image Caching Bug

Coolify has a known bug ([#5587](https://github.com/coollabsio/coolify/issues/5587), [#7084](https://github.com/coollabsio/coolify/issues/7084)) where docker-compose images pulled from registries are cached and not re-pulled on deploy, even when the remote `:main` tag has been updated.

**Workaround:** The docker-compose uses `image: ghcr.io/r1ft1/bonk/backend:${TAG:-main}`. The GitHub Actions workflow sets the `TAG` environment variable in Coolify to the commit SHA before triggering deploy. Since each SHA is a unique tag, Docker has no cached version and must pull fresh.

The workflow updates the env var via the Coolify API (PATCH doesn't work for envs, so it does delete + create):
1. `GET /api/v1/applications/<uuid>/envs` — find TAG env UUID
2. `DELETE /api/v1/applications/<uuid>/envs/<tag-uuid>` — remove old value
3. `POST /api/v1/applications/<uuid>/envs` — create with new SHA value
4. `GET <webhook-url>` — trigger deploy

### Build Caching

GitHub Actions uses `cache-from: type=gha` / `cache-to: type=gha,mode=max` with separate scopes for frontend and backend. This caches intermediate Docker layers between builds — if only source code changed, the `npm ci` / `go mod download` layers are reused. Typically saves 30-50% build time.

### GitHub Actions Secrets Required

| Secret | Purpose |
|---|---|
| `COOLIFY_WEBHOOK` | Deploy webhook URL |
| `COOLIFY_TOKEN` | Coolify API bearer token |
| `COOLIFY_API_URL` | Coolify API base (e.g. `https://coolify.oatmocha.com/api/v1`) |
| `COOLIFY_APP_UUID` | Application UUID in Coolify |

### Local Development

```bash
colima start                    # Start Docker runtime
docker compose -f docker-compose.dev.yaml up -d
# Frontend: http://localhost:5173 (hot reload, source mounted)
# Backend: http://localhost:8080 (needs --build to pick up Go changes)
```

To rebuild backend after Go changes:
```bash
docker compose -f docker-compose.dev.yaml up -d --build backend
```

## Traefik + Coolify Routing

### The Problem

Coolify auto-generates Traefik labels (`http-0-*`, `https-0-*`) for docker-compose apps. These conflict with custom labels in the compose file, causing Traefik error: "Router cannot be linked automatically with multiple Services."

### The Fix

Every custom router label MUST explicitly specify which service to use:

```yaml
- "traefik.http.routers.boop-frontend.service=boop-frontend"
- "traefik.http.services.boop-frontend.loadbalancer.server.port=3000"
```

Without the explicit `.service=` link, Traefik sees multiple services (yours + Coolify's auto-generated ones) and refuses to route.

### Known Pitfalls

- **`pull_policy: always`** in docker-compose breaks Coolify deploys — do NOT use
- Coolify's auto-generated router names (`http-0-*`) change between deploys — never reference them
- For middleware, use `coolify.traefik.middlewares` label instead of manual router middleware labels
- If HTTPS stops working (504), check if Traefik proxy needs a restart (`docker restart coolify-proxy` on VPS) — stale ACME/TLS state can block routing
- `$` in compose labels must be escaped as `$$` to avoid Docker variable interpolation

### Troubleshooting 504s

1. Check container is running: `docker ps | grep yg4gkks`
2. Check container logs: `docker logs <container-name>`
3. Test internal connectivity: `docker exec coolify-proxy wget -q -O- http://<container>:3000`
4. Check Traefik errors: `docker logs coolify-proxy --tail 20 | grep ERR`
5. If Traefik can reach container internally but external 504: restart proxy `docker restart coolify-proxy`

## Docker Image Security

Both images run as non-root user `appuser` (uid 1001). The backend's `/data` directory is owned by `appuser` for SQLite and log file write access. Go binary is statically compiled (`CGO_ENABLED=0`) with stripped symbols (`-ldflags="-s -w"`).
