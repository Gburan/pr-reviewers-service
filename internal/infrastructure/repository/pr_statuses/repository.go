package pr_statuses

import (
	"context"
	"errors"
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
	prStatusesTableName = "pr_statuses"
	idColumnName        = "id"
	statusColumnName    = "status"

	returnAll = "RETURNING *"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{db: pool}
}

func (r *Repository) SavePRStatus(ctx context.Context, status PRStatusIn) (*PRStatusOut, error) {
	if status.ID == uuid.Nil {
		status.ID = uuid.New()
	}

	queryBuilder := squirrel.Insert(prStatusesTableName).
		PlaceholderFormat(squirrel.Dollar).
		Columns(idColumnName, statusColumnName).
		Values(status.ID, status.Status).
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

	_, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[prStatusDB])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	slog.DebugContext(ctx, "Repository SavePRStatus success")
	return &PRStatusOut{
		ID:     status.ID,
		Status: status.Status,
	}, nil
}

func (r *Repository) GetPRStatusByID(ctx context.Context, status PRStatusIn) (*PRStatusOut, error) {
	selectBuilder := squirrel.
		Select(idColumnName, statusColumnName).
		PlaceholderFormat(squirrel.Dollar).
		From(prStatusesTableName).
		Where(squirrel.Eq{idColumnName: status.ID})

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

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[prStatusDB])
	if err != nil {
		slog.DebugContext(ctx, err.Error())
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %v", repository.ErrPRStatusNotFound, err)
		}
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	slog.DebugContext(ctx, "Repository GetPRStatusByID success")
	return &PRStatusOut{
		ID:     result.ID,
		Status: result.Status,
	}, nil
}

func (r *Repository) GetPRStatusesByIDs(ctx context.Context, statusIDs []uuid.UUID) (*[]PRStatusOut, error) {
	if len(statusIDs) == 0 {
		return &[]PRStatusOut{}, nil
	}

	selectBuilder := squirrel.
		Select(idColumnName, statusColumnName).
		PlaceholderFormat(squirrel.Dollar).
		From(prStatusesTableName).
		Where(squirrel.Eq{idColumnName: statusIDs})

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

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[prStatusDB])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	statuses := make([]PRStatusOut, 0, len(results))
	for _, result := range results {
		statuses = append(statuses, PRStatusOut(result))
	}

	slog.DebugContext(ctx, "Repository GetPRStatusesByIDs success")
	return &statuses, nil
}

func (r *Repository) UpdatePRStatusByID(ctx context.Context, status PRStatusIn) (*PRStatusOut, error) {
	queryBuilder := squirrel.Update(prStatusesTableName).
		PlaceholderFormat(squirrel.Dollar).
		Set(statusColumnName, status.Status).
		Where(squirrel.Eq{idColumnName: status.ID}).
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

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[prStatusDB])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	slog.DebugContext(ctx, "Repository UpdatePRStatus success")
	return &PRStatusOut{
		ID:     result.ID,
		Status: result.Status,
	}, nil
}
