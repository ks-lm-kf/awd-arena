package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ContainerCPUPercent = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "awd_container_cpu_percent",
		Help: "Container CPU usage percentage",
	}, []string{"team_id", "container_id"})

	ContainerMemoryBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "awd_container_memory_bytes",
		Help: "Container memory usage in bytes",
	}, []string{"team_id", "container_id"})

	AttackTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "awd_attack_total",
		Help: "Total number of attack submissions",
	}, []string{"game_id", "attacker_team"})

	AttackSuccessTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "awd_attack_success_total",
		Help: "Total number of successful attacks",
	}, []string{"game_id", "attacker_team"})

	RoundDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "awd_round_duration_seconds",
		Help:    "Round processing duration in seconds",
		Buckets: prometheus.DefBuckets,
	})

	FlagSubmitTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "awd_flag_submit_total",
		Help: "Total flag submissions",
	}, []string{"game_id", "team_id"})

	ActiveContainers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "awd_active_containers",
		Help: "Number of currently active containers",
	})
)
