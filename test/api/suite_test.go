package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"pr-reviewers-service/internal/app"
	suite2 "pr-reviewers-service/test/suite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	migrationsDir = "../../migrations"
)

type ApiTest struct {
	suite2.TestSuite
	App *app.App
}

func (s *ApiTest) SetupSuite() {
	s.InitConfig()
	suite2.Config.DB.MigrationsDir = migrationsDir
	suite2.Config.Server.Rest.Address = ":31337"

	var err error
	s.Container, err = s.InitDB()
	assert.NoError(s.T(), err)

	ctx := s.withSkipMigrations(context.Background())
	s.App, err = app.NewApp(ctx, suite2.Config)
	assert.NoError(s.T(), err)

	err = s.GetTables(suite2.GlobalPool, ctx)
	assert.NoError(s.T(), err)

	go func() {
		if err = s.App.RestRun(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.T().Errorf("rest error: %v", err)
		}
	}()
	//go func() {
	//	if err = s.App.GrpcRun(); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
	//		s.T().Errorf("grpc error: %v", err)
	//	}
	//}()

	time.Sleep(2 * time.Second)
}

func (s *ApiTest) withSkipMigrations(ctx context.Context) context.Context {
	return context.WithValue(ctx, app.SkipMigrationsKey, true)
}

func (s *ApiTest) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.App.Stop(ctx); err != nil {
		s.T().Errorf("error stop app: %v", err)
	}
}

func (s *ApiTest) SetupTest() {
	ctx := context.Background()
	truncateSQL := fmt.Sprintf("%s %s %s", "TRUNCATE TABLE", strings.Join(s.Tables, ", "), "CASCADE;")
	_, err := suite2.GlobalPool.Exec(ctx, truncateSQL)
	assert.NoError(s.T(), err)
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(ApiTest))
}
