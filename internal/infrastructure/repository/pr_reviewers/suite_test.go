package pr_reviewers

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

type PRReviewersTest struct {
	suite2.TestSuite
}

func (s *PRReviewersTest) SetupSuite() {
	s.InitConfig()
	suite2.Config.DB.MigrationsDir = migrationsDir

	var err error
	s.Container, err = s.InitDB()
	assert.NoError(s.T(), err)

	ctx := context.Background()
	err = s.GetTables(suite2.GlobalPool, ctx)
	assert.NoError(s.T(), err)
}

func (s *PRReviewersTest) SetupTest() {
	ctx := context.Background()
	truncateSQL := fmt.Sprintf("%s %s %s", "TRUNCATE TABLE", strings.Join(s.Tables, ", "), "CASCADE;")
	_, err := suite2.GlobalPool.Exec(ctx, truncateSQL)
	assert.NoError(s.T(), err)
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(PRReviewersTest))
}
