package pull_requests

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"pr-reviewers-service/internal/infrastructure/repository"
	nower2 "pr-reviewers-service/internal/usecase/contract/nower"

	"github.com/Masterminds/squirrel"
	trm "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	pullRequestsTableName = "pull_requests"
	idColumnName          = "id"
	nameColumnName        = "name"
	authorIdColumnName    = "author_id"
	statusIdColumnName    = "status_id"
	createdAtColumnName   = "created_at"
	mergedAtColumnName    = "merged_at"

	returnAll = "RETURNING *"
)

type Repository struct {
	db    *pgxpool.Pool
	nower nower2.Nower
}

func NewRepository(pool *pgxpool.Pool, nower nower2.Nower) *Repository {
	return &Repository{db: pool, nower: nower}
}

func (r *Repository) SavePullRequest(ctx context.Context, pr PullRequestIn) (*PullRequestOut, error) {
	if pr.ID == uuid.Nil {
		pr.ID = uuid.New()
	}
	if pr.CreatedAt.IsZero() {
		pr.CreatedAt = r.nower.Now()
	}

	queryBuilder := squirrel.Insert(pullRequestsTableName).
		PlaceholderFormat(squirrel.Dollar).
		Columns(idColumnName, nameColumnName, authorIdColumnName, statusIdColumnName, createdAtColumnName, mergedAtColumnName).
		Values(pr.ID, pr.Name, pr.AuthorID, pr.StatusID, pr.CreatedAt, pr.MergedAt).
		Suffix(returnAll)

	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrBuildQuery, err)
	}

	q := trm.DefaultCtxGetter.DefaultTrOrDB(ctx, r.db)
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrExecuteQuery, err)
	}
	defer rows.Close()

	_, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[pullRequestDB])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	slog.DebugContext(ctx, "Repository SavePullRequest success")
	return &PullRequestOut{
		ID:        pr.ID,
		Name:      pr.Name,
		AuthorID:  pr.AuthorID,
		StatusID:  pr.StatusID,
		CreatedAt: pr.CreatedAt,
		MergedAt:  pr.MergedAt,
	}, nil
}

func (r *Repository) GetPullRequestByID(ctx context.Context, prID uuid.UUID) (*PullRequestOut, error) {
	selectBuilder := squirrel.
		Select(idColumnName, nameColumnName, authorIdColumnName, statusIdColumnName, createdAtColumnName, mergedAtColumnName).
		PlaceholderFormat(squirrel.Dollar).
		From(pullRequestsTableName).
		Where(squirrel.Eq{idColumnName: prID})

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrBuildQuery, err)
	}

	q := trm.DefaultCtxGetter.DefaultTrOrDB(ctx, r.db)
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrExecuteQuery, err)
	}
	defer rows.Close()

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[pullRequestDB])
	if err != nil {
		slog.DebugContext(ctx, err.Error())
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %v", repository.ErrPullRequestNotFound, err)
		}
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	slog.DebugContext(ctx, "Repository GetPullRequestByID success")
	return &PullRequestOut{
		ID:        result.ID,
		Name:      result.Name,
		AuthorID:  result.AuthorID,
		StatusID:  result.StatusID,
		CreatedAt: result.CreatedAt,
		MergedAt:  result.MergedAt,
	}, nil
}

func (r *Repository) GetPullRequestsByPrIDs(ctx context.Context, prIDs []uuid.UUID) (*[]PullRequestOut, error) {
	if len(prIDs) == 0 {
		return &[]PullRequestOut{}, nil
	}

	selectBuilder := squirrel.
		Select(idColumnName, nameColumnName, authorIdColumnName, statusIdColumnName, createdAtColumnName, mergedAtColumnName).
		PlaceholderFormat(squirrel.Dollar).
		From(pullRequestsTableName).
		Where(squirrel.Eq{idColumnName: prIDs})

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrBuildQuery, err)
	}

	q := trm.DefaultCtxGetter.DefaultTrOrDB(ctx, r.db)
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrExecuteQuery, err)
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[pullRequestDB])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	prs := make([]PullRequestOut, 0, len(results))
	for _, result := range results {
		prs = append(prs, PullRequestOut(result))
	}

	slog.DebugContext(ctx, "Repository GetPullRequestsByPrIDs success", "count", len(prs))
	return &prs, nil
}

func (r *Repository) MarkPullRequestMergedByID(ctx context.Context, prID uuid.UUID) (*PullRequestOut, error) {
	now := r.nower.Now()

	queryBuilder := squirrel.Update(pullRequestsTableName).
		PlaceholderFormat(squirrel.Dollar).
		Set(mergedAtColumnName, now).
		Where(squirrel.Eq{idColumnName: prID}).
		Suffix(returnAll)

	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrBuildQuery, err)
	}

	q := trm.DefaultCtxGetter.DefaultTrOrDB(ctx, r.db)
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrExecuteQuery, err)
	}
	defer rows.Close()

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[pullRequestDB])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	slog.DebugContext(ctx, "Repository UpdatePullRequest success")
	return &PullRequestOut{
		ID:        result.ID,
		Name:      result.Name,
		AuthorID:  result.AuthorID,
		StatusID:  result.StatusID,
		CreatedAt: result.CreatedAt,
		MergedAt:  result.MergedAt,
	}, nil
}
