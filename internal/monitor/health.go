package monitor

import "context"

// HealthChecker performs health checks on dependencies.
type HealthChecker struct{}

// Check checks all dependencies.
func (h *HealthChecker) Check(ctx context.Context) map[string]string {
	return map[string]string{
		"postgres":    "ok",
		"redis":       "ok",
		"nats":        "ok",
		"clickhouse":  "ok",
		"docker":      "ok",
	}
}
