# AWD Arena Core Refactoring Progress

## Completed Changes (Build Verified)

### P0-1: Container Auto Management
- **Fixed team querying**: `ContainerService.ProvisionContainers()` now uses `GameTeam` table to get game-specific teams instead of fetching ALL teams from the database
- **Port mapping persistence**: `CreateChallengeContainer()` now properly serializes port mappings to JSON and saves to `TeamContainer.PortMapping` field
- **Engine cleanup on stop**: `CompetitionEngine.Stop()` now calls `ContainerService.TeardownContainers()` to clean up all game containers when the engine stops
- **SSH credentials**: Already implemented - each container gets random SSH password, SSH user `awd` is created via exec
- **Container lifecycle**: Already fully implemented - create, start, pause, unpause, restart, destroy all working

### P0-2: Flag Auto Rotation
- **Fixed team querying**: `FlagService.GenerateFlags()` now uses `GameTeam` table for game-specific teams
- **Unified flag format**: Changed to `flag{gameID_round_teamID_challengeID_random}` for consistency
- **Flag writing**: `CompetitionEngine.onRoundStart()` generates flags via `FlagService.GenerateFlags()` and writes to containers via `FlagWriter.WriteFlag()` (writes to `/flag` file)
- **Flag hash storage**: Flags are hashed with SHA256 and stored in `FlagRecord` table

### P0-3: Player Machine Info API
- **Added route**: `GET /api/v1/games/:id/my-machines` (aliases existing `/my-containers`)
- **Returns**: Container info including IP, SSH user/password, port mappings, challenge name, status
- **Auth**: Requires JWT + `PermViewOwnStats` permission (player role)
- **Implementation**: `ScoreService.GetUserContainers()` looks up user's team, queries `TeamContainer`, enriches with challenge names

### P0-4: Flag Submit Verification
- **Endpoint**: `POST /api/v1/games/:id/flags/submit` (already implemented)
- **Flow**: Hash comparison against `FlagRecord` table -> correct: record `FlagSubmission`, award points, broadcast leaderboard update / wrong: record failed submission
- **Scoring**: `BaseAttackPoints * AttackWeight`, updates team cumulative score
- **WebSocket**: Broadcasts `flag:captured` event on correct submission
- **Dedup**: Prevents same attacker from submitting same flag twice
- **Self-flag**: Prevents submitting own team's flag

### Scoring System
- **Fixed**: `ScoreCalculator.CalculateRoundScores()` now uses `GameTeam` table
- **Round scoring**: Attack scores (captured flags) and defense losses
- **Cumulative scores**: Sum of all round scores + adjustments
- **Ranking updates**: Automatic rank calculation after each round

## Files Modified
- `internal/service/container_service.go` - GameTeam-based team queries, port mapping fix
- `internal/service/flag_service.go` - GameTeam-based team queries, unified flag format
- `internal/engine/engine.go` - Container cleanup on stop
- `internal/engine/scoring.go` - GameTeam-based team queries
- `internal/container/manager.go` - Port mapping JSON serialization
- `internal/server/router.go` - Added /my-machines route

## Build Status
- All changes compile successfully with `go build ./cmd/server/`

