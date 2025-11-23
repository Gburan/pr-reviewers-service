package teams

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
	teamsTableName      = "teams"
	idColumnName        = "id"
	nameColumnName      = "name"
	createdAtColumnName = "created_at"

	returnAll = "RETURNING *"
)

type Repository struct {
	db    *pgxpool.Pool
	nower nower2.Nower
}

func NewRepository(pool *pgxpool.Pool, nower nower2.Nower) *Repository {
	return &Repository{db: pool, nower: nower}
}

func (r *Repository) SaveTeam(ctx context.Context, team TeamIn) (*TeamOut, error) {
	if team.ID == uuid.Nil {
		team.ID = uuid.New()
	}
	if team.CreatedAt.IsZero() {
		team.CreatedAt = r.nower.Now()
	}

	queryBuilder := squirrel.Insert(teamsTableName).
		PlaceholderFormat(squirrel.Dollar).
		Columns(idColumnName, nameColumnName, createdAtColumnName).
		Values(team.ID, team.Name, team.CreatedAt).
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

	_, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[teamDB])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	slog.DebugContext(ctx, "Repository SaveTeam success")
	return &TeamOut{
		ID:        team.ID,
		Name:      team.Name,
		CreatedAt: team.CreatedAt,
	}, nil
}

func (r *Repository) GetTeamByID(ctx context.Context, teamId uuid.UUID) (*TeamOut, error) {
	selectBuilder := squirrel.
		Select(idColumnName, nameColumnName, createdAtColumnName).
		PlaceholderFormat(squirrel.Dollar).
		From(teamsTableName).
		Where(squirrel.Eq{idColumnName: teamId})

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

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[teamDB])
	if err != nil {
		slog.DebugContext(ctx, err.Error())
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %v", repository.ErrTeamNotFound, err)
		}
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	slog.DebugContext(ctx, "Repository GetTeamByID success")
	return &TeamOut{
		ID:        result.ID,
		Name:      result.Name,
		CreatedAt: result.CreatedAt,
	}, nil
}

func (r *Repository) GetTeamByName(ctx context.Context, name string) (*TeamOut, error) {
	selectBuilder := squirrel.
		Select(idColumnName, nameColumnName, createdAtColumnName).
		PlaceholderFormat(squirrel.Dollar).
		From(teamsTableName).
		Where(squirrel.Eq{nameColumnName: name})

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

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[teamDB])
	if err != nil {
		slog.DebugContext(ctx, err.Error())
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %v", repository.ErrTeamNotFound, err)
		}
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	slog.DebugContext(ctx, "Repository GetTeamByName success")
	return &TeamOut{
		ID:        result.ID,
		Name:      result.Name,
		CreatedAt: result.CreatedAt,
	}, nil
}
