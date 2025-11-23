package users

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
	usersTableName      = "users"
	idColumnName        = "id"
	nameColumnName      = "name"
	isActiveColumnName  = "is_active"
	teamIdColumnName    = "team_id"
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

func (r *Repository) GetUserByID(ctx context.Context, userId uuid.UUID) (*UserOut, error) {
	selectBuilder := squirrel.
		Select(idColumnName, nameColumnName, isActiveColumnName, teamIdColumnName, createdAtColumnName).
		PlaceholderFormat(squirrel.Dollar).
		From(usersTableName).
		Where(squirrel.Eq{idColumnName: userId})

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

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userDB])
	if err != nil {
		slog.DebugContext(ctx, err.Error())
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %v", repository.ErrUserNotFound, err)
		}
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	slog.DebugContext(ctx, "Repository GetUserByID success")
	return &UserOut{
		ID:        result.ID,
		Name:      result.Name,
		IsActive:  result.IsActive,
		TeamID:    result.TeamID,
		CreatedAt: result.CreatedAt,
	}, nil
}

func (r *Repository) GetUsersByIDs(ctx context.Context, userIds []uuid.UUID) (*[]UserOut, error) {
	if len(userIds) == 0 {
		return &[]UserOut{}, nil
	}

	selectBuilder := squirrel.
		Select(idColumnName, nameColumnName, isActiveColumnName, teamIdColumnName, createdAtColumnName).
		PlaceholderFormat(squirrel.Dollar).
		From(usersTableName).
		Where(squirrel.Eq{idColumnName: userIds})

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

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[userDB])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	users := make([]UserOut, 0, len(results))
	for _, result := range results {
		users = append(users, UserOut{
			ID:        result.ID,
			Name:      result.Name,
			IsActive:  result.IsActive,
			TeamID:    result.TeamID,
			CreatedAt: result.CreatedAt,
		})
	}

	slog.DebugContext(ctx, "Repository GetUsersByIDs success")
	return &users, nil
}

func (r *Repository) GetUsersByTeamID(ctx context.Context, teamID uuid.UUID) (*[]UserOut, error) {
	selectBuilder := squirrel.
		Select(idColumnName, nameColumnName, isActiveColumnName, teamIdColumnName, createdAtColumnName).
		PlaceholderFormat(squirrel.Dollar).
		From(usersTableName).
		Where(squirrel.Eq{teamIdColumnName: teamID})

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

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[userDB])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	users := make([]UserOut, 0, len(results))
	for _, result := range results {
		users = append(users, UserOut{
			ID:        result.ID,
			Name:      result.Name,
			IsActive:  result.IsActive,
			TeamID:    result.TeamID,
			CreatedAt: result.CreatedAt,
		})
	}

	slog.DebugContext(ctx, "Repository GetUsersByTeamID success")
	return &users, nil
}

func (r *Repository) GetActiveUsersByTeamID(ctx context.Context, teamID uuid.UUID) (*[]UserOut, error) {
	selectBuilder := squirrel.
		Select(idColumnName, nameColumnName, isActiveColumnName, teamIdColumnName, createdAtColumnName).
		PlaceholderFormat(squirrel.Dollar).
		From(usersTableName).
		Where(squirrel.Eq{teamIdColumnName: teamID, isActiveColumnName: true})

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

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[userDB])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	users := make([]UserOut, 0, len(results))
	for _, result := range results {
		users = append(users, UserOut{
			ID:        result.ID,
			Name:      result.Name,
			IsActive:  result.IsActive,
			TeamID:    result.TeamID,
			CreatedAt: result.CreatedAt,
		})
	}

	slog.DebugContext(ctx, "Repository GetActiveUsersByTeamID success")
	return &users, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user UserIn) (*UserOut, error) {
	queryBuilder := squirrel.Update(usersTableName).
		PlaceholderFormat(squirrel.Dollar).
		Set(nameColumnName, user.Name).
		Set(isActiveColumnName, user.IsActive).
		Set(teamIdColumnName, user.TeamID).
		Where(squirrel.Eq{idColumnName: user.ID}).
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

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userDB])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	slog.DebugContext(ctx, "Repository UpdateUser success")
	return &UserOut{
		ID:        result.ID,
		Name:      result.Name,
		IsActive:  result.IsActive,
		TeamID:    result.TeamID,
		CreatedAt: result.CreatedAt,
	}, nil
}

func (r *Repository) SaveUsersBatch(ctx context.Context, users []UserIn) (*[]UserOut, error) {
	if len(users) == 0 {
		return &[]UserOut{}, nil
	}

	queryBuilder := squirrel.Insert(usersTableName).
		PlaceholderFormat(squirrel.Dollar).
		Columns(idColumnName, nameColumnName, isActiveColumnName, teamIdColumnName, createdAtColumnName)

	now := r.nower.Now()
	for _, user := range users {
		userID := user.ID
		if userID == uuid.Nil {
			userID = uuid.New()
		}

		queryBuilder = queryBuilder.Values(userID, user.Name, user.IsActive, user.TeamID, now)
	}
	queryBuilder = queryBuilder.Suffix(returnAll)

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

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[userDB])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	userOuts := make([]UserOut, 0, len(results))
	for _, result := range results {
		userOuts = append(userOuts, UserOut{
			ID:        result.ID,
			Name:      result.Name,
			IsActive:  result.IsActive,
			TeamID:    result.TeamID,
			CreatedAt: result.CreatedAt,
		})
	}

	return &userOuts, nil
}

func (r *Repository) UpdateUsersBatch(ctx context.Context, users []UserIn) (*[]UserOut, error) {
	if len(users) == 0 {
		return &[]UserOut{}, nil
	}

	ids := make([]interface{}, len(users))
	names := make([]interface{}, len(users))
	isActives := make([]interface{}, len(users))
	teamIDs := make([]interface{}, len(users))
	for i, user := range users {
		ids[i] = user.ID
		names[i] = user.Name
		isActives[i] = user.IsActive
		teamIDs[i] = user.TeamID
	}

	dataTable := squirrel.
		Select().
		Column(fmt.Sprintf("unnest(?::uuid[]) AS %s", idColumnName), ids).
		Column(fmt.Sprintf("unnest(?::text[]) AS %s", nameColumnName), names).
		Column(fmt.Sprintf("unnest(?::boolean[]) AS %s", isActiveColumnName), isActives).
		Column(fmt.Sprintf("unnest(?::uuid[]) AS %s", teamIdColumnName), teamIDs)
	queryBuilder := squirrel.Update(usersTableName).
		PlaceholderFormat(squirrel.Dollar).
		Set(nameColumnName, squirrel.Expr(fmt.Sprintf("data_table.%s", nameColumnName))).
		Set(isActiveColumnName, squirrel.Expr(fmt.Sprintf("data_table.%s", isActiveColumnName))).
		Set(teamIdColumnName, squirrel.Expr(fmt.Sprintf("data_table.%s", teamIdColumnName))).
		FromSelect(dataTable, "data_table").
		Where(fmt.Sprintf("%s.%s = data_table.%s", usersTableName, idColumnName, idColumnName)).
		Suffix(fmt.Sprintf("RETURNING %s.*", usersTableName))

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

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[userDB])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, fmt.Errorf("%w: %v", repository.ErrScanResult, err)
	}

	userOuts := make([]UserOut, 0, len(results))
	for _, result := range results {
		userOuts = append(userOuts, UserOut{
			ID:        result.ID,
			Name:      result.Name,
			IsActive:  result.IsActive,
			TeamID:    result.TeamID,
			CreatedAt: result.CreatedAt,
		})
	}

	slog.DebugContext(ctx, "Repository UpdateUsersBatch success", "count", len(userOuts))
	return &userOuts, nil
}
