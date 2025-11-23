package suite

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"pr-reviewers-service/internal/config"

	"github.com/Masterminds/squirrel"
	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbImage = "postgres:16-alpine"

	logLevel = "info"
	cfgPath  = "./config/config.yaml"
)

var (
	excludeTables = []string{"goose_db_version"}

	DBOnce sync.Once

	Config     config.Config
	ConfigOnce sync.Once

	GlobalPool *pgxpool.Pool
)

type TestSuite struct {
	suite.Suite
	Tables    []string
	Container testcontainers.Container
}

func (s *TestSuite) InitConfig() {
	findProjectRoot := func() string {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		for {
			modPath := filepath.Join(dir, "go.mod")
			if _, err := os.Stat(modPath); err == nil {
				return dir
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				slog.Error("go.mod not found", "error", err)
				os.Exit(1)
			}
			dir = parent
		}
	}

	ConfigOnce.Do(func() {
		Config = config.MustLoad(filepath.Join(findProjectRoot(), cfgPath))
		Config.App.Logging.Level = logLevel
	})
}

func (s *TestSuite) InitDB() (testcontainers.Container, error) {
	slog.Debug("Current DB connection string:", "conn", Config.DB.Conn)
	slog.Debug("Migrations directory:", "dir", Config.DB.MigrationsDir)

	info, err := os.Stat(Config.DB.MigrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Error("Migrations directory does not exist:", "dir", Config.DB.MigrationsDir)
			os.Exit(1)
		} else {
			slog.Error("Error accessing migrations directory:", "error", err)
			os.Exit(1)
		}
	} else if !info.IsDir() {
		slog.Error("Migrations path is not a directory:", "dir", Config.DB.MigrationsDir)
		os.Exit(1)
	}

	DBOnce.Do(func() {
		containerSetup := func(ctx context.Context) {
			u, err := url.Parse(Config.DB.Conn)
			if err != nil {
				slog.Error("url parse from cfg:", "error", err)
				os.Exit(1)
			}
			password, _ := u.User.Password()
			if err != nil {
				slog.Error("url parse from cfg:", "error", err)
				os.Exit(1)
			}
			_, port, err := net.SplitHostPort(u.Host)
			if err != nil {
				slog.Error("url parse from cfg:", "error", err)
				os.Exit(1)
			}
			var env = map[string]string{
				"POSTGRES_PASSWORD": password,
				"POSTGRES_USER":     u.User.Username(),
				"POSTGRES_DB":       strings.TrimPrefix(u.Path, "/"),
			}
			req := testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Image:        dbImage,
					ExposedPorts: []string{fmt.Sprintf("%s/tcp", port)},
					Env:          env,
					WaitingFor:   wait.ForLog("database system is ready to accept connections"),
					Name:         "test-postgres",
				},
				Started: true,
				Reuse:   true,
			}
			container, err := testcontainers.GenericContainer(ctx, req)
			if err != nil {
				slog.Error("failed to start container:", "error", err)
				os.Exit(1)
			}
			mappedPort, err := container.MappedPort(ctx, nat.Port(port))
			if err != nil {
				slog.Error("failed to get container external port:", "error", err)
				os.Exit(1)
			}
			hostname := u.Hostname()
			u.Host = fmt.Sprintf("%s:%s", hostname, mappedPort.Port())
			Config.DB.Conn = u.String()

			s.Container = container

			time.Sleep(2 * time.Second)
			slog.Info("postgres container ready and running at port: ", "port", mappedPort.Port())
		}
		migrationsSetup := func(ctx context.Context) {
			dsn := flag.String("dsn", Config.DB.Conn, "PostgreSQL")
			sql, err := goose.OpenDBWithDriver("postgres", *dsn)
			if err != nil {
				slog.Error("failed to OpenDBWithDriver:", "error", err)
				os.Exit(1)
			}
			if err = goose.Up(sql, Config.DB.MigrationsDir); err != nil {
				slog.Error("failed to migrate db:", "error", err)
				os.Exit(1)
			}
		}
		poolSetup := func(ctx context.Context) {
			cfgDb, err := pgxpool.ParseConfig(Config.DB.Conn)
			if err != nil {
				slog.Error("failed to parse config:", "error", err)
				os.Exit(1)
			}
			GlobalPool, err = pgxpool.NewWithConfig(context.Background(), cfgDb)
			if err != nil {
				slog.Error("failed to init pool:", "error", err)
				os.Exit(1)
			}
		}

		funcs := []func(context.Context){
			containerSetup,
			poolSetup,
			migrationsSetup,
		}
		ctx := context.Background()
		for _, f := range funcs {
			f(ctx)
		}
	})

	return s.Container, nil
}

func (s *TestSuite) GetTables(pool *pgxpool.Pool, ctx context.Context) error {
	selectBuilder := squirrel.
		Select("tablename").
		PlaceholderFormat(squirrel.Dollar).
		From("pg_tables").
		Where(squirrel.Eq{"schemaname": "public"})

	if len(excludeTables) > 0 {
		args := make([]any, len(excludeTables))
		for i, v := range excludeTables {
			args[i] = v
		}

		placeholders := strings.Repeat("?,", len(excludeTables))
		placeholders = placeholders[:len(placeholders)-1]

		selectBuilder = selectBuilder.Where(
			fmt.Sprintf("tablename NOT IN (%s)", placeholders),
			args...,
		)
	}

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return err
	}

	rows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err = rows.Scan(&tableName); err != nil {
			return err
		}
		tables = append(tables, tableName)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	s.Tables = tables
	return nil
}
