package script

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"pg-start-trainee-2024/domain/entity"
)

type Repo struct {
	DB *sqlx.DB
}

func New(db *sqlx.DB) *Repo {
	return &Repo{
		DB: db,
	}
}

func (r *Repo) CreateScript(ctx context.Context, script entity.Script) (*entity.Script, error) {
	result, err := r.DB.NamedQueryContext(ctx,
		`INSERT INTO script (command, output, is_running, pid) VALUES (:command, :output, :is_running, :pid) 
RETURNING id, command, output, is_running, pid, created_at, updated_at`,
		&script)
	if err != nil {
		return nil, err
	}

	if result.Next() {
		if err = result.StructScan(&script); err != nil {
			return nil, err
		}
	}

	return &script, nil
}

func (r *Repo) queryRowxContextWithStructScan(ctx context.Context, query string, dest any) error {
	result := r.DB.QueryRowxContext(ctx, query)

	if err := result.Err(); err != nil {
		return err
	}

	if err := result.StructScan(dest); err != nil {
		return err
	}

	return nil
}

func (r *Repo) UpdateScriptOutput(ctx context.Context, id int, output string) (*entity.Script, error) {
	var script entity.Script

	// TODO: bad concatenation!
	if err := r.queryRowxContextWithStructScan(
		ctx,
		fmt.Sprintf(`UPDATE script SET output = output || '%v' WHERE id = %v 
        RETURNING  id, command, output, is_running, pid, created_at, updated_at`, output, id),
		&script,
	); err != nil {
		return nil, err
	}

	return &script, nil
}

func (r *Repo) DeleteScript(ctx context.Context, id int) (*entity.Script, error) {
	var script entity.Script

	if err := r.queryRowxContextWithStructScan(
		ctx,
		fmt.Sprintf(`DELETE FROM script WHERE id = %v
        RETURNING  id, command, output, is_running, pid, created_at, updated_at`, id),
		&script,
	); err != nil {
		return nil, err
	}

	return &script, nil
}

func (r *Repo) UpdateScriptPID(ctx context.Context, id, pid int) (*entity.Script, error) {
	var script entity.Script

	if err := r.queryRowxContextWithStructScan(
		ctx,
		fmt.Sprintf(`UPDATE script SET pid = %v WHERE id = %v
        RETURNING  id, command, output, is_running, pid, created_at, updated_at`, pid, id),
		&script,
	); err != nil {
		return nil, err
	}

	return &script, nil
}
