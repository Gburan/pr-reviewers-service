package pr_reviewers

import (
	"context"
	"fmt"
	"log/slog"

	"pr-reviewers-service/internal/infrastructure/repository"

	"github.com/Masterminds/squirrel"
	trm "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	prReviewersTableName = "pr_reviewers"
	idColumnName         = "id"
	prIdColumnName       = "pr_id"
	reviewerIdColumnName = "reviewer_id"

	returnAll = "RETURNING *"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{db: pool}
}

func (r *Repository) SavePRReviewer(ctx context.Context, reviewer PrReviewerIn) (*PrReviewerOut, error) {
	if reviewer.ID == uuid.Nil {
		reviewer.ID = uuid.New()
	}

	queryBuilder := squirrel.Insert(prReviewersTableName).
		PlaceholderFormat(squirrel.Dollar).
		Columns(idColumnName, prIdColumnName, reviewerIdColumnName).
		Values(reviewer.ID, reviewer.PrID, reviewer.ReviewerID).
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

	_, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[prReviewerDB])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	slog.DebugContext(ctx, "Repository SavePRReviewer success")
	return &PrReviewerOut{
		ID:         reviewer.ID,
		PRID:       reviewer.PrID,
		ReviewerID: reviewer.ReviewerID,
	}, nil
}

func (r *Repository) GetPRReviewersByPRID(ctx context.Context, prID uuid.UUID) (*[]PrReviewerOut, error) {
	selectBuilder := squirrel.
		Select(idColumnName, prIdColumnName, reviewerIdColumnName).
		PlaceholderFormat(squirrel.Dollar).
		From(prReviewersTableName).
		Where(squirrel.Eq{prIdColumnName: prID})

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

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[prReviewerDB])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	reviewers := make([]PrReviewerOut, 0, len(results))
	for _, result := range results {
		reviewers = append(reviewers, PrReviewerOut(result))
	}

	slog.DebugContext(ctx, "Repository GetPRReviewersByPRID success")
	return &reviewers, nil
}

func (r *Repository) GetPRReviewersByReviewerID(ctx context.Context, reviewerID uuid.UUID) (*[]PrReviewerOut, error) {
	selectBuilder := squirrel.
		Select(idColumnName, prIdColumnName, reviewerIdColumnName).
		PlaceholderFormat(squirrel.Dollar).
		From(prReviewersTableName).
		Where(squirrel.Eq{reviewerIdColumnName: reviewerID})

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

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[prReviewerDB])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	reviewers := make([]PrReviewerOut, 0, len(results))
	for _, result := range results {
		reviewers = append(reviewers, PrReviewerOut(result))
	}

	slog.DebugContext(ctx, "Repository GetPRReviewersByReviewerID success")
	return &reviewers, nil
}

func (r *Repository) GetPRReviewersByReviewerIDs(ctx context.Context, reviewerIDs []uuid.UUID) (*[]PrReviewerOut, error) {
	if len(reviewerIDs) == 0 {
		slog.DebugContext(ctx, "Repository GetPRReviewersByReviewerIDs: empty reviewer IDs list")
		return &[]PrReviewerOut{}, nil
	}

	selectBuilder := squirrel.
		Select(idColumnName, prIdColumnName, reviewerIdColumnName).
		PlaceholderFormat(squirrel.Dollar).
		From(prReviewersTableName).
		Where(squirrel.Eq{reviewerIdColumnName: reviewerIDs})

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		slog.ErrorContext(ctx, "Repository GetPRReviewersByReviewerIDs: build query error", "error", err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrBuildQuery, err)
	}

	q := trm.DefaultCtxGetter.DefaultTrOrDB(ctx, r.db)
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		slog.ErrorContext(ctx, "Repository GetPRReviewersByReviewerIDs: execute query error", "error", err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrExecuteQuery, err)
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[prReviewerDB])
	if err != nil {
		slog.ErrorContext(ctx, "Repository GetPRReviewersByReviewerIDs: scan results error", "error", err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	reviewers := make([]PrReviewerOut, 0, len(results))
	for _, result := range results {
		reviewers = append(reviewers, PrReviewerOut(result))
	}

	slog.DebugContext(ctx, "Repository GetPRReviewersByReviewerIDs success",
		"reviewer_ids_count", len(reviewerIDs),
		"found_reviewers_count", len(reviewers))
	return &reviewers, nil
}

func (r *Repository) GetAllPRReviewers(ctx context.Context) (*[]PrReviewerOut, error) {
	selectBuilder := squirrel.
		Select(idColumnName, prIdColumnName, reviewerIdColumnName).
		PlaceholderFormat(squirrel.Dollar).
		From(prReviewersTableName)

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

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[prReviewerDB])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	reviewers := make([]PrReviewerOut, 0, len(results))
	for _, result := range results {
		reviewers = append(reviewers, PrReviewerOut(result))
	}

	slog.DebugContext(ctx, "Repository GetAllPRReviewers success", "count", len(reviewers))
	return &reviewers, nil
}

func (r *Repository) DeletePRReviewerByPRAndReviewer(ctx context.Context, prID, reviewerID uuid.UUID) error {
	queryBuilder := squirrel.Delete(prReviewersTableName).
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{prIdColumnName: prID, reviewerIdColumnName: reviewerID})

	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return fmt.Errorf("%w: %v", repository.ErrBuildQuery, err)
	}

	q := trm.DefaultCtxGetter.DefaultTrOrDB(ctx, r.db)
	_, err = q.Exec(ctx, sql, args...)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return fmt.Errorf("%w: %v", repository.ErrExecuteQuery, err)
	}

	slog.DebugContext(ctx, "Repository DeletePRReviewerByPRAndReviewer success")
	return nil
}
