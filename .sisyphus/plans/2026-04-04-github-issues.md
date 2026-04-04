# GitHub Issues Fix Plan — 2026-04-04 08:00 CST

## Summary

19 open issues. 2 already fixed in commit `9feb86e`. 17 remaining.

### Already Fixed (close on GitHub tomorrow)

| # | Issue | How Fixed |
|---|-------|-----------|
| #1 | README 默认密码不一致 | README updated to `Admin@2026` |
| #12 | 审计日志路由前后端不一致 | Backend audit routes moved to `/api/v1/audit/` |

### Remaining 17 Issues — Priority Order

## Batch 1: Backend Data/API Bugs (8 issues, do FIRST)

### #2 — 排行榜字段名不匹配 (HIGH)
- **File**: `internal/service/ranking_service.go`, `internal/handler/ranking_handler.go`
- **Fix**: Backend returns `score` → rename to `total_score`; `defense_loss` → rename to `defense_score`; add `flag_count` field
- **Verify**: Frontend `web/src/types/index.ts` RankingItem expects `total_score`, `defense_score`, `flag_count`

### #3 — AddTeamToGame 返回零值 team (HIGH)
- **File**: `internal/handler/admin_handler.go` → `AddTeamToGame()`
- **Fix**: Add `db.Preload("Team")` after creating the GameTeam record, then re-query
- **Verify**: Response team object should have real data

### #4 — 分数调整返回全0 (HIGH)
- **File**: `internal/handler/admin_handler.go` → `AdjustScore()`
- **Fix**: Bind request body correctly, actually update score in DB, return old/new scores
- **Verify**: POST adjust with adjustment=100 → returns non-zero values

### #13 — 列表接口泄漏 raw_token (HIGH, security)
- **File**: `internal/handler/team_handler.go` → `List()`, `Get()`
- **Fix**: In List/Get responses, omit `raw_token` field. Only return it in Create response
- **Verify**: GET /teams should NOT contain raw_token in any item

### #8 — 注册不验证密码强度 (MEDIUM)
- **File**: `internal/service/auth_service.go` → `Register()`
- **Fix**: Call `validatePasswordStrength(password, username)` before hashing
- **Verify**: POST register with password "123456" → 400 error

### #10 — 创建用户返回500而非具体错误 (MEDIUM)
- **File**: `internal/handler/user_handler.go` → `Create()`
- **Fix**: Distinguish errors — duplicate username → 400 "username already exists"; password=username → 400; other → 500
- **Verify**: Duplicate username → 400 with message

### #14 — 空比赛可启动 (MEDIUM)
- **File**: `internal/handler/game_handler.go` → `Start()` or `internal/engine/game_state_machine.go`
- **Fix**: Before state transition to "running", check: game has ≥1 team AND ≥1 challenge
- **Verify**: Start empty game → 400 "比赛至少需要一支队伍和一道题目"

### #15 — 比赛停止后可提交Flag (MEDIUM)
- **File**: `internal/service/flag_service.go` → `SubmitFlag()` or `internal/handler/flag_handler.go`
- **Fix**: Check game state is "running" before accepting flag. If not → 400 "比赛未在进行中"
- **Verify**: Submit flag after game stopped → 400 with message

## Batch 2: Backend + Frontend Coordination (4 issues)

### #16 — Settings 返回 msg 而非 message (LOW)
- **File**: `internal/handler/settings_handler.go`
- **Fix**: Change all `c.JSON(fiber.Map{"msg": ...})` to `"message"`
- **Verify**: GET /settings → response has `message` field

### #11 — Challenge exposed_ports 类型不匹配 (MEDIUM)
- **File**: `internal/model/challenge.go` + handler/service, `web/src/types/index.ts`
- **Fix**: Backend should store ports as JSON, return as array. Or align both to string format
- **Verify**: Frontend can render port info correctly

### #6 — 无队伍口令加入接口 (MEDIUM, new feature)
- **File**: New route in `router.go`, new handler method, `internal/service/team_service.go`
- **Fix**: Add `POST /api/v1/teams/join` — validate token against team raw_token (hashed comparison), set user.team_id
- **Frontend**: Wire register page token field to call this API after registration
- **Verify**: Register with team token → user joins team

### #19 — 新用户 must_change_password 默认 false (MEDIUM)
- **File**: `internal/service/auth_service.go` → `Register()`
- **Fix**: Set `MustChangePassword: true` on new user creation
- **Verify**: Register new user → must_change_password is true in DB

## Batch 3: Frontend Bugs (3 issues)

### #7 — admin首次登录不引导改密码 (HIGH)
- **File**: `web/src/App.tsx` or router guards, `web/src/api/client.ts`
- **Fix**: 
  1. Fix `/api/v1/auth/me` to return actual `must_change_password` from DB (backend fix in auth_handler.go Me())
  2. Frontend: intercept 403 with "password change required" → redirect to change-password page
- **Verify**: Admin first login → auto-redirect to password change

### #17 — TypeScript 编译错误 (LOW)
- **File**: `web/src/pages/admin/ContainerManage.tsx`, `web/src/pages/admin/GameManage.tsx`
- **Fix**: 
  1. ContainerManage: fix `container_id` type mismatch
  2. GameManage: change `status === "active"` to `status === "running"`
- **Verify**: `npx tsc --noEmit` passes

### #18 — 排行榜重复空状态 (LOW)
- **File**: `web/src/pages/Ranking/index.tsx` or similar
- **Fix**: Remove duplicate empty state. Keep table "暂无数据" only, OR hide table and show single empty state
- **Verify**: No data → only ONE empty state shown

## Batch 4: Architecture / Cleanup (2 issues)

### #5 — 模板管理页面404 (LOW)
- **Note**: Backend template_handler.go was removed (dead code). Frontend pages still exist.
- **Fix**: Remove frontend template pages and routes from `web/src/App.tsx`
- **Files to remove**: `web/src/pages/Template/` or similar, API file, route entry
- **Verify**: /templates → 404 or redirect (not error toast)

### #9 — 两套重复管理页面 (MEDIUM, refactor)
- **File**: `web/src/App.tsx`, `web/src/pages/`
- **Fix**: Merge `/games` and `/admin/games` into one. Use role-based visibility for buttons. Remove duplicate routes.
- **This is the biggest task** — may span many files. Consider deferring if time-constrained.

## Execution Strategy

1. **8:00-8:30** — Batch 1 (#2, #3, #4, #13, #8, #10, #14, #15) — Backend only, fast fixes
2. **8:30-9:00** — Batch 2 (#16, #11, #6, #19) — Backend + minor frontend
3. **9:00-9:30** — Batch 3 (#7, #17, #18) — Frontend fixes
4. **9:30-9:45** — Batch 4 (#5) — Remove dead frontend pages
5. **9:45-10:00** — Close all fixed issues on GitHub, commit, push
6. **#9 deferred** — Requires significant refactoring, create separate plan

## Commands to Run Tomorrow

```bash
# Start fresh session, then reference this plan:
# "读取 .sisyphus/plans/2026-04-04-github-issues.md 并按计划修复所有issue"

# Verify build after each batch:
go build ./cmd/server/

# Close fixed issues:
gh issue close 1 --comment "Fixed in commit 9feb86e"
gh issue close 12 --comment "Fixed in commit 9feb86e"
# ... close others as fixed

# Push when done:
git push origin master
```
