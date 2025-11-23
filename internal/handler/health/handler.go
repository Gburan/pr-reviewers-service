package health

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/hellofresh/health-go/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	errHealthCheckFailed = errors.New("health check failed")
)

type HealthChecker struct {
	health *health.Health
}

func NewHealthChecker(pool *pgxpool.Pool, serviceName, version string) (*HealthChecker, error) {
	h, err := health.New(
		health.WithComponent(health.Component{
			Name:    serviceName,
			Version: version,
		}),
		health.WithSystemInfo(),
		health.WithChecks(
			health.Config{
				Name:    "postgres",
				Timeout: 5 * time.Second,
				Check:   createPostgresCheck(pool),
			},
		),
	)
	if err != nil {
		return nil, err
	}

	return &HealthChecker{health: h}, nil
}

func (hc *HealthChecker) Handler() http.Handler {
	return hc.health.Handler()
}

// HealthCheck godoc
// @Summary Health check
// @Description Check service health and DB connection
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Service health"
// @Failure 503 {object} map[string]interface{} "Service unhealth"
// @Router /health [get]
func (hc *HealthChecker) HandlerFunc() http.HandlerFunc {
	return hc.health.HandlerFunc
}

func createPostgresCheck(pool *pgxpool.Pool) health.CheckFunc {
	return func(ctx context.Context) error {
		conn, err := pool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		var result int
		err = conn.QueryRow(ctx, "SELECT 1").Scan(&result)
		if err != nil {
			return err
		}

		if result != 1 {
			return errHealthCheckFailed
		}

		return nil
	}
}
