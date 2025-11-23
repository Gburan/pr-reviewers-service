package api

import (
	"net/http"

	"pr-reviewers-service/internal/generated/api/v1/handler"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func (s *ApiTest) TestAllAPIHelpers() {
	loginResp, status, _, err := dummyLogin()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, status)
	token := loginResp.Token
	teamName := "TestTeam"
	user1ID := uuid.New()
	user2ID := uuid.New()
	user3ID := uuid.New()
	user4ID := uuid.New()
	members := []handler.TeamMember{
		{
			UserId:   user1ID,
			Username: "user1",
			IsActive: true,
		},
		{
			UserId:   user2ID,
			Username: "user2",
			IsActive: true,
		},
		{
			UserId:   user3ID,
			Username: "user3",
			IsActive: true,
		},
		{
			UserId:   user4ID,
			Username: "user4",
			IsActive: true,
		},
	}

	team, status, _, err := addTeam(token, teamName, members)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, status)

	prID := uuid.New()
	prName := "Test PR"

	prResp, status, _, err := createPullRequest(token, user1ID, prID, prName)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, status)

	statsResp, status, _, err := getReviewersStats(token)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, status)

	teamResp, status, _, err := getTeam(token, teamName)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, status)

	userReviewResp, status, _, err := getUserReviewPRs(token, user2ID.String())
	assert.NoError(s.T(), err)
	assert.True(s.T(), status == http.StatusOK || status == http.StatusNotFound, "Expected 200 or 404, got %d", status)

	oldReviewerID := prResp.Pr.AssignedReviewers[0]
	reassignResp, status, _, err := reassignPullRequest(token, prID, oldReviewerID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, status)
	assert.NotNil(s.T(), reassignResp)

	mergeResp, status, _, err := mergePullRequest(token, prID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, status)

	deactivateResp, status, _, err := deactivateTeamUsers(token, teamName, []uuid.UUID{user3ID})
	assert.NoError(s.T(), err)
	assert.True(s.T(), status == http.StatusOK || status == http.StatusNotFound, "Expected 200 or 404, got %d", status)

	setActiveResp, status, _, err := setUserActiveStatus(token, user3ID, false)
	assert.NoError(s.T(), err)
	assert.True(s.T(), status == http.StatusOK || status == http.StatusNotModified, "Expected 200 or 304, got %d", status)

	assert.NotNil(s.T(), loginResp)
	assert.NotNil(s.T(), team)
	assert.NotNil(s.T(), prResp)
	assert.NotNil(s.T(), statsResp)
	assert.NotNil(s.T(), teamResp)
	if status == http.StatusOK {
		assert.NotNil(s.T(), userReviewResp)
	}
	assert.NotNil(s.T(), mergeResp)
	if status == http.StatusOK {
		assert.NotNil(s.T(), deactivateResp)
	}
	if status == http.StatusOK {
		assert.NotNil(s.T(), setActiveResp)
	}
}
