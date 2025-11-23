package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"reflect"

	_ "pr-reviewers-service/docs/rest"
	add_team2 "pr-reviewers-service/internal/handler/add_team"
	"pr-reviewers-service/internal/handler/dummy_login"
	get_review2 "pr-reviewers-service/internal/handler/get_review"
	get_team2 "pr-reviewers-service/internal/handler/get_team"
	"pr-reviewers-service/internal/handler/health"
	"pr-reviewers-service/internal/handler/middleware"
	pull_request_create2 "pr-reviewers-service/internal/handler/pull_request_create"
	pull_request_merge2 "pr-reviewers-service/internal/handler/pull_request_merge"
	pull_request_reassign2 "pr-reviewers-service/internal/handler/pull_request_reassign"
	set_is_active2 "pr-reviewers-service/internal/handler/set_is_active"
	stats_pr_assignments2 "pr-reviewers-service/internal/handler/stats_pr_assignments"
	team_deactivate_users2 "pr-reviewers-service/internal/handler/team_deactivate_users"
	nower2 "pr-reviewers-service/internal/infrastructure/nower"
	randomizer2 "pr-reviewers-service/internal/infrastructure/randomizer"
	"pr-reviewers-service/internal/infrastructure/repository/pr_reviewers"
	"pr-reviewers-service/internal/infrastructure/repository/pr_statuses"
	"pr-reviewers-service/internal/infrastructure/repository/pull_requests"
	"pr-reviewers-service/internal/infrastructure/repository/teams"
	"pr-reviewers-service/internal/infrastructure/repository/users"
	"pr-reviewers-service/internal/logging"
	"pr-reviewers-service/internal/metrics"
	"pr-reviewers-service/internal/usecase/add_team"
	"pr-reviewers-service/internal/usecase/get_review"
	"pr-reviewers-service/internal/usecase/get_team"
	"pr-reviewers-service/internal/usecase/pull_request_create"
	"pr-reviewers-service/internal/usecase/pull_request_merge"
	"pr-reviewers-service/internal/usecase/pull_request_reassign"
	"pr-reviewers-service/internal/usecase/set_is_active"
	"pr-reviewers-service/internal/usecase/stats_pr_assignments"
	"pr-reviewers-service/internal/usecase/team_deactivate_users"

	trmpgxv5 "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

type ctxKey string

const SkipMigrationsKey ctxKey = "skipMigrations"

var (
	// nolint:unused
	userRoleOnly = []middleware.UserRole{middleware.User}
	// nolint:unused
	adminRoleOnly = []middleware.UserRole{middleware.Admin}
	allRoles      = []middleware.UserRole{middleware.User, middleware.Admin}
)

func (a *App) setup(ctx context.Context) error {
	funcs := []func(context.Context) error{
		a.setupLogger,
		a.setupMetrics,
		a.setupValidator,
		a.setupDbPoolTrManager,
		a.setupRestServer,
		a.setupGrpcServer,
		a.setupMigrationsDB,
	}

	for _, f := range funcs {
		if err := f(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) setupLogger(_ context.Context) error {
	var output io.Writer
	switch a.config.App.Logging.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		if err := os.MkdirAll(filepath.Dir(a.config.App.Logging.Output), 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		f, err := os.OpenFile(a.config.App.Logging.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		output = f
	}

	var level slog.Level
	switch a.config.App.Logging.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.Handler(slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level: level,
	}))
	handler = logging.NewLoggerImpl(handler)
	slog.SetDefault(slog.New(handler))

	return nil
}

func (a *App) setupMetrics(_ context.Context) error {
	prometheus.MustRegister(
		metrics.RestRequestsTotal,
		metrics.RestResponseDuration,
		metrics.RestEndpointsResponsesTotal,
		metrics.CreatedTeams,
		metrics.CreatedUsers,
		metrics.CreatedPRs,
	)
	return nil
}

// @title           PR Reviewers service
// @version         1.0
// @description     This is a service for working with PR Reviewers.

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey  ApiKeyAuth
// @in header
// @name Authorization
func (a *App) setupRestServer(_ context.Context) error {
	nower := nower2.Nower{}
	randomizer := randomizer2.Randomizer{}

	healthChecker, err := health.NewHealthChecker(a.pool, "pr-reviewer-service", "1.0")
	if err != nil {
		return err
	}

	repPrReviewers := pr_reviewers.NewRepository(a.pool)
	repPrStatuses := pr_statuses.NewRepository(a.pool)
	repPullRequests := pull_requests.NewRepository(a.pool, nower)
	repTeams := teams.NewRepository(a.pool, nower)
	repUsers := users.NewRepository(a.pool, nower)

	dummy := dummy_login.New(a.config.App.JWTSecret, a.validator)
	addTeamUseCase := add_team.Newusecase(repUsers, repTeams, a.trManager)
	addTeam := add_team2.New(addTeamUseCase, a.validator)
	getTeamUsecase := get_team.NewUsecase(repTeams, repUsers)
	getTeam := get_team2.New(getTeamUsecase)

	setIsActiveUseCase := set_is_active.NewUsecase(repTeams, repUsers, a.trManager)
	setIsActive := set_is_active2.New(setIsActiveUseCase, a.validator)
	getReviewUseCase := get_review.NewUsecase(repUsers, repPullRequests, repPrReviewers, repPrStatuses)
	getReview := get_review2.New(getReviewUseCase, a.validator)

	prCreateUseCase := pull_request_create.NewUsecase(repUsers, repPullRequests, repPrReviewers,
		repPrStatuses, randomizer, a.config.App.Validation.MaxPrReviewers, a.trManager)
	prCreate := pull_request_create2.New(prCreateUseCase, a.validator)
	prMergeUseCase := pull_request_merge.NewUsecase(repPullRequests, repPrReviewers, repPrStatuses, a.trManager)
	prMerge := pull_request_merge2.New(prMergeUseCase, a.validator)
	reassignUseCase := pull_request_reassign.NewUsecase(repUsers, repPullRequests, repPrReviewers,
		repPrStatuses, randomizer, a.config.App.Validation.MaxPrReviewers, a.trManager)
	reassign := pull_request_reassign2.New(reassignUseCase, a.validator)

	statsPrAssignmentsUseCase := stats_pr_assignments.NewUsecase(repPrReviewers)
	stats := stats_pr_assignments2.New(statsPrAssignmentsUseCase)

	deactivateTeamUseCase := team_deactivate_users.NewUsecase(repTeams, repUsers, repPullRequests,
		repPrReviewers, repPrStatuses, randomizer, a.trManager)
	deactivateTeam := team_deactivate_users2.New(deactivateTeamUseCase, a.validator)

	middlewares := func(mustBeOneOfRole []middleware.UserRole, h http.HandlerFunc) http.Handler {
		handler := h
		if a.config.App.AuthorisationNeeded && len(mustBeOneOfRole) != 0 {
			handler = middleware.AuthMiddleware(a.config.App.JWTSecret, mustBeOneOfRole, handler)
		}
		handler = middleware.LoggerMiddleware(handler)
		handler = middleware.PanicMiddleware(handler)
		return handler
	}

	r := mux.NewRouter()
	v1 := r.PathPrefix("/api/v1").Subrouter()
	r.Handle("/health", healthChecker.HandlerFunc()).Methods("GET")
	r.Handle("/metrics", promhttp.Handler()).Methods("GET")
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	v1.Handle("/dummyLogin", middlewares(nil, dummy.DummyLogin)).Methods("POST")

	teamV1 := v1.PathPrefix("/team").Subrouter()
	teamV1.Handle("/add", middlewares(allRoles, addTeam.AddTeam)).Methods("POST")
	teamV1.Handle("/get", middlewares(allRoles, getTeam.GetTeam)).Methods("GET")
	teamV1.Handle("/deactivateUsers", middlewares(allRoles, deactivateTeam.DeactivateTeamUsers)).Methods("PATCH")

	usersV1 := v1.PathPrefix("/users").Subrouter()
	usersV1.Handle("/setIsActive", middlewares(allRoles, setIsActive.SetIsActive)).Methods("POST")
	usersV1.Handle("/getReview", middlewares(allRoles, getReview.GetUserReviewPRs)).Methods("GET")

	prV1 := v1.PathPrefix("/pullRequest").Subrouter()
	prV1.Handle("/create", middlewares(allRoles, prCreate.CreatePullRequest)).Methods("POST")
	prV1.Handle("/merge", middlewares(allRoles, prMerge.MergePullRequest)).Methods("POST")
	prV1.Handle("/reassign", middlewares(allRoles, reassign.ReassignPullRequest)).Methods("POST")

	statV1 := v1.PathPrefix("/statistics").Subrouter()
	statV1.Handle("/reviewers", middlewares(allRoles, stats.GetReviewersStats)).Methods("GET")

	a.restServer = &http.Server{
		Addr:         a.config.Server.Rest.Address,
		ReadTimeout:  a.config.Server.Rest.Connsettings.ReadTimeout,
		WriteTimeout: a.config.Server.Rest.Connsettings.WriteTimeout,
		IdleTimeout:  a.config.Server.Rest.Connsettings.IdleTimeout,
		Handler:      r,
	}

	return nil
}

func (a *App) setupGrpcServer(_ context.Context) error {
	//grpcServer := grpc.NewServer(
	//	grpc.KeepaliveParams(keepalive.ServerParameters{
	//		MaxConnectionIdle: a.config.Server.GRPC.ConnSettings.MaxConnIdle,
	//		MaxConnectionAge:  a.config.Server.GRPC.ConnSettings.MaxConnAge,
	//	}),
	//)
	//
	//nower := nower2.Nower{}
	//repPVZ := pvz.NewRepository(a.pool, nower)
	//listpvzsUsecase := list_pvzs.NewUsecase(repPVZ)
	//pvzServer := server.New(listpvzsUsecase)
	//
	//pvz_v1.RegisterPVZServiceServer(grpcServer, pvzServer)
	//a.grpcServer = grpcServer

	return nil
}

func (a *App) setupDbPoolTrManager(ctx context.Context) error {
	cfg, err := pgxpool.ParseConfig(a.config.DB.Conn)
	if err != nil {
		return err
	}
	cfg.MaxConns = a.config.DB.PoolSettings.MaxConns
	cfg.MaxConnIdleTime = a.config.DB.PoolSettings.MaxConnIdleTime
	cfg.MinIdleConns = a.config.DB.PoolSettings.MinIdleConns
	cfg.MaxConnLifetime = a.config.DB.PoolSettings.MaxConnLifetime
	a.pool, err = pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return err
	}

	a.trManager = manager.Must(trmpgxv5.NewDefaultFactory(a.pool))

	return nil
}

func (a *App) setupMigrationsDB(ctx context.Context) error {
	if skip, ok := ctx.Value(SkipMigrationsKey).(bool); ok && skip {
		return nil
	}

	dsn := flag.String("dsn", a.config.DB.Conn, "PostgreSQL")
	sql, err := goose.OpenDBWithDriver("postgres", *dsn)
	if err != nil {
		return err
	}
	if err = goose.Up(sql, a.config.DB.MigrationsDir); err != nil {
		return err
	}

	return nil
}

func (a *App) setupValidator(_ context.Context) error {
	a.validator = validator.New()
	registerOneOf := func(category string, allowed []string) error {
		return a.validator.RegisterValidation(category, func(fl validator.FieldLevel) bool {
			value := fl.Field().String()
			for _, allow := range allowed {
				if value == allow {
					return true
				}
			}
			return false
		})
	}
	registerNumeric := func(category string, maxValue int) error {
		return a.validator.RegisterValidation(category, func(fl validator.FieldLevel) bool {
			field := fl.Field()

			switch field.Kind() {
			case reflect.Slice, reflect.Array:
				return field.Len() <= maxValue
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return field.Int() <= int64(maxValue)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return field.Uint() <= uint64(maxValue)
			default:
				return false
			}
		})
	}

	validationsOneOf := map[string][]string{
		"oneof_user": a.config.App.Validation.AllowedUsers,
	}
	for cat, catAllow := range validationsOneOf {
		err := registerOneOf(cat, catAllow)
		if err != nil {
			return fmt.Errorf("registering oneof category %s error: %v", cat, err)
		}
	}

	numericValidations := map[string]int{
		"max_pr_reviewers": a.config.App.Validation.MaxPrReviewers,
	}
	for cat, maxVal := range numericValidations {
		err := registerNumeric(cat, maxVal)
		if err != nil {
			return fmt.Errorf("registering numeric category %s error: %v", cat, err)
		}
	}

	return nil
}
