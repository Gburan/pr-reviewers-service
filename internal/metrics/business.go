package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	CreatedPRs = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "created_prs_total",
			Help: "Total number of created Pull Requests.",
		},
	)

	CreatedTeams = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "created_teams_total",
			Help: "Total number of created teams.",
		},
	)

	CreatedUsers = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "created_users_total",
			Help: "Total number of created users.",
		},
	)
)

func IncCreatedPRs() {
	CreatedPRs.Inc()
}

func IncCreatedTeams() {
	CreatedTeams.Inc()
}

func IncCreatedUsers(cnt int) {
	CreatedUsers.Add(float64(cnt))
}
