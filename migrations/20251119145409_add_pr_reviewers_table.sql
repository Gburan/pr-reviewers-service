-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL,
    team_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS pull_requests (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    author_id UUID NOT NULL,
    status_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    merged_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS pr_statuses (
    id UUID PRIMARY KEY,
    status VARCHAR(20) NOT NULL
);

CREATE TABLE IF NOT EXISTS pr_reviewers (
    id UUID PRIMARY KEY,
    pr_id UUID NOT NULL,
    reviewer_id UUID NOT NULL
);

ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_team_id;
ALTER TABLE pull_requests DROP CONSTRAINT IF EXISTS fk_pull_requests_author_id;
ALTER TABLE pull_requests DROP CONSTRAINT IF EXISTS fk_pull_requests_status_id;
ALTER TABLE pr_reviewers DROP CONSTRAINT IF EXISTS fk_pr_reviewers_pr_id;
ALTER TABLE pr_reviewers DROP CONSTRAINT IF EXISTS fk_pr_reviewers_reviewer_id;

ALTER TABLE users ADD CONSTRAINT fk_users_team_id FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;
ALTER TABLE pull_requests ADD CONSTRAINT fk_pull_requests_author_id FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE pull_requests ADD CONSTRAINT fk_pull_requests_status_id FOREIGN KEY (status_id) REFERENCES pr_statuses(id);
ALTER TABLE pr_reviewers ADD CONSTRAINT fk_pr_reviewers_pr_id FOREIGN KEY (pr_id) REFERENCES pull_requests(id) ON DELETE CASCADE;
ALTER TABLE pr_reviewers ADD CONSTRAINT fk_pr_reviewers_reviewer_id FOREIGN KEY (reviewer_id) REFERENCES users(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE pr_reviewers DROP CONSTRAINT IF EXISTS fk_pr_reviewers_reviewer_id;
ALTER TABLE pr_reviewers DROP CONSTRAINT IF EXISTS fk_pr_reviewers_pr_id;
ALTER TABLE pull_requests DROP CONSTRAINT IF EXISTS fk_pull_requests_status_id;
ALTER TABLE pull_requests DROP CONSTRAINT IF EXISTS fk_pull_requests_author_id;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_team_id;

DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pr_statuses;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
-- +goose StatementEnd
