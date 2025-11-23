package teams

import (
	"context"
	"fmt"
	"strings"
	"testing"

	suite2 "pr-reviewers-service/test/suite"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	migrationsDir = "../../../../migrations/"
)

type TeamsTest struct {
	suite2.TestSuite
}

func (s *TeamsTest) SetupSuite() {
	s.InitConfig()
	suite2.Config.DB.MigrationsDir = migrationsDir

	var err error
	s.Container, err = s.InitDB()
	assert.NoError(s.T(), err)

	ctx := context.Background()
	err = s.GetTables(suite2.GlobalPool, ctx)
	assert.NoError(s.T(), err)
}

func (s *TeamsTest) SetupTest() {
	ctx := context.Background()
	truncateSQL := fmt.Sprintf("%s %s %s", "TRUNCATE TABLE", strings.Join(s.Tables, ", "), "CASCADE;")
	_, err := suite2.GlobalPool.Exec(ctx, truncateSQL)
	assert.NoError(s.T(), err)
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(TeamsTest))
}
