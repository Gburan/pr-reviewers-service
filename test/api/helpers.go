package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"pr-reviewers-service/internal/generated/api/v1/handler"

	"github.com/google/uuid"
)

const (
	testServerAddr = "http://localhost:31337"

	methodPost  = "POST"
	methodGet   = "GET"
	methodPatch = "PATCH"

	apiV1 = "/api/v1"
)

func apiPathV1(path string) string {
	return fmt.Sprintf("%s%s", apiV1, path)
}

func dummyLogin() (handler.DummyLoginOut, int, handler.ErrorResponse, error) {
	return doRequest[struct{}, handler.DummyLoginOut](methodPost, apiPathV1("/dummyLogin"), "")
}

func createPullRequest(token string, authorID, prID uuid.UUID, prName string) (handler.CreatePullRequestResponse, int, handler.ErrorResponse, error) {
	in := handler.PostPullRequestCreateJSONRequestBody{
		AuthorId:        authorID,
		PullRequestId:   prID,
		PullRequestName: prName,
	}
	return doRequest[handler.PostPullRequestCreateJSONRequestBody, handler.CreatePullRequestResponse](methodPost, apiPathV1("/pullRequest/create"), token, in)
}

func mergePullRequest(token string, prID uuid.UUID) (handler.MergePullRequestResponse, int, handler.ErrorResponse, error) {
	in := handler.PostPullRequestMergeJSONRequestBody{
		PullRequestId: prID,
	}
	return doRequest[handler.PostPullRequestMergeJSONRequestBody, handler.MergePullRequestResponse](methodPost, apiPathV1("/pullRequest/merge"), token, in)
}

func reassignPullRequest(token string, prID, oldReviewerID uuid.UUID) (handler.ReassignPullRequestResponse, int, handler.ErrorResponse, error) {
	in := handler.PostPullRequestReassignJSONRequestBody{
		PullRequestId: prID,
		OldReviewerId: oldReviewerID,
	}
	return doRequest[handler.PostPullRequestReassignJSONRequestBody, handler.ReassignPullRequestResponse](methodPost, apiPathV1("/pullRequest/reassign"), token, in)
}

func getReviewersStats(token string) (handler.ReviewersStatsResponse, int, handler.ErrorResponse, error) {
	return doRequest[struct{}, handler.ReviewersStatsResponse](methodGet, apiPathV1("/statistics/reviewers"), token)
}

func addTeam(token, teamName string, members []handler.TeamMember) (handler.Team, int, handler.ErrorResponse, error) {
	in := handler.PostTeamAddJSONRequestBody{
		TeamName: teamName,
		Members:  members,
	}
	return doRequest[handler.PostTeamAddJSONRequestBody, handler.Team](methodPost, apiPathV1("/team/add"), token, in)
}

func deactivateTeamUsers(token, teamName string, userIDs []uuid.UUID) (handler.DeactivateTeamUsersResponse, int, handler.ErrorResponse, error) {
	in := handler.PatchTeamDeactivateUsersJSONRequestBody{
		TeamName: teamName,
		UserIds:  userIDs,
	}
	return doRequest[handler.PatchTeamDeactivateUsersJSONRequestBody, handler.DeactivateTeamUsersResponse](methodPatch, apiPathV1("/team/deactivateUsers"), token, in)
}

func getTeam(token, teamName string) (handler.Team, int, handler.ErrorResponse, error) {
	encodedTeamName := url.QueryEscape(teamName)
	url := fmt.Sprintf("/team/get?team_name=%s", encodedTeamName)
	return doRequest[struct{}, handler.Team](methodGet, apiPathV1(url), token)
}

func getUserReviewPRs(token, userID string) (handler.GetUserReviewPRsResponse, int, handler.ErrorResponse, error) {
	encodedUserID := url.QueryEscape(userID)
	url := fmt.Sprintf("/users/getReview?user_id=%s", encodedUserID)
	return doRequest[struct{}, handler.GetUserReviewPRsResponse](methodGet, apiPathV1(url), token)
}

func setUserActiveStatus(token string, userID uuid.UUID, isActive bool) (handler.SetUserActiveStatusResponse, int, handler.ErrorResponse, error) {
	in := handler.PostUsersSetIsActiveJSONRequestBody{
		UserId:   userID,
		IsActive: isActive,
	}
	return doRequest[handler.PostUsersSetIsActiveJSONRequestBody, handler.SetUserActiveStatusResponse](methodPost, apiPathV1("/users/setIsActive"), token, in)
}

func doRequest[T, K any](method, path, token string, inData ...T) (K, int, handler.ErrorResponse, error) {
	var result K
	var body io.Reader

	if len(inData) > 0 {
		jsonBody, err := json.Marshal(inData[0])
		if err != nil {
			return result, 0, handler.ErrorResponse{}, err
		}
		body = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", testServerAddr, path), body)
	if err != nil {
		return result, 0, handler.ErrorResponse{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return result, 0, handler.ErrorResponse{}, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, resp.StatusCode, handler.ErrorResponse{}, err
	}

	if resp.StatusCode >= 400 {
		var errResp handler.ErrorResponse
		if err = json.Unmarshal(respBody, &errResp); err == nil {
			return result, resp.StatusCode, errResp, nil
		}
		return result, resp.StatusCode, handler.ErrorResponse{}, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	if err = json.Unmarshal(respBody, &result); err != nil {
		return result, resp.StatusCode, handler.ErrorResponse{}, err
	}

	return result, resp.StatusCode, handler.ErrorResponse{}, nil
}
