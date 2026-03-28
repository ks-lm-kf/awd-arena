# AWD Arena Core Refactoring Progress

## Completed Changes (Build Verified)

### P0-1: Container Auto Management
- GameTeam-based team queries in ContainerService
- Port mapping persistence in JSON
- Engine cleanup on stop via TeardownContainers
- SSH credentials per container

### P0-2: Flag Auto Rotation
- GameTeam-based team queries in FlagService
- Unified flag format: `flag{gameID_round_teamID_challengeID_random}`
- Flag writing to `/flag` in containers
- SHA256 hash storage

### P0-3: Player Machine Info API
- `GET /api/v1/games/:id/my-machines` endpoint
- Returns IP, SSH, ports, challenge info
- JWT auth with PermViewOwnStats

### P0-4: Flag Submit Verification
- Hash comparison against FlagRecord
- Scoring with BaseAttackPoints * AttackWeight
- Dedup and self-flag prevention
- WebSocket broadcast on capture

### P2-1: Image Management (Complete)
- `GET /api/v1/admin/images/host/list` - list host Docker images
- `POST /api/v1/admin/images/pull` - pull from Docker Hub
- `POST /api/v1/admin/images/` - create image record
- DockerImages frontend page with host list, pull, build, CRUD
- Challenge creation can reference existing images

### P2-2: Game Export (Complete)
- `GET /api/v1/games/:id/export/scoreboard/csv` - ranking CSV
- `GET /api/v1/games/:id/export/scoreboard/pdf` - ranking HTML
- `GET /api/v1/games/:id/export/attacks` - attack log CSV
- `GET /api/v1/games/:id/export/all` - all export links
- All handlers use `database.GetDB()` (not c.Locals("db"))

### P2-3: Frontend Pages (Complete)
- **AttackPanel**: Shows player machines (IP/port/SSH), submit flag, history
- **Ranking**: Real-time ranking with WebSocket, round selection
- **GameManage (admin)**: CRUD games, start/pause/stop via judge API
- **ContainerManage (admin)**: View all containers, restart individual/bulk
- **GameDetail**: Full game management with teams, challenges, rankings
- **DockerImages**: Image management with host list, pull, build
- All pages use antd components, react-query, zustand auth store

### P2-4: E2E Test Script
- Written to `/tmp/e2e_test.py`
- Tests: login, create game/teams/challenge, start, rankings, machines, exports, stop
- Run: `python3 /tmp/e2e_test.py`

## Files Modified in P2
- `internal/handler/leaderboard_handler.go` - Rewritten to use database.GetDB()
- `internal/handler/auth_handler.go` - Fixed c.Locals("db") -> database.GetDB()
- `web/src/pages/AttackPanel/index.tsx` - Machine info, flag submit, history
- `internal/handler/export_handler.go` - Already correct (uses database.GetDB())

## Build Status
- Go backend: Builds successfully
- React frontend: Builds successfully (vite + tsc)
