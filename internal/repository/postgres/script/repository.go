package script

import (
	"context"
	"fmt"
	"math"

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

	defer result.Close()

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

	return result.StructScan(dest)
}

func (r *Repo) UpdateScriptOutput(ctx context.Context, id int, output string) (*entity.Script, error) {
	var script entity.Script

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

func (r *Repo) UpdateScriptPIDAndRunningState(ctx context.Context, id, pid int, isRunning bool) (*entity.Script, error) {
	var script entity.Script

	if err := r.queryRowxContextWithStructScan(
		ctx,
		fmt.Sprintf(`UPDATE script SET pid = %v, is_running = %v WHERE id = %v
        RETURNING  id, command, output, is_running, pid, created_at, updated_at`, pid, isRunning, id),
		&script,
	); err != nil {
		return nil, err
	}

	return &script, nil
}

func (r *Repo) UpdateScriptRunningState(ctx context.Context, id int, isRunning bool) (*entity.Script, error) {
	var script entity.Script

	if err := r.queryRowxContextWithStructScan(
		ctx,
		fmt.Sprintf(`UPDATE script SET is_running = %v WHERE id = %v
        RETURNING  id, command, output, is_running, pid, created_at, updated_at`, isRunning, id),
		&script,
	); err != nil {
		return nil, err
	}

	return &script, nil
}

func (r *Repo) GetScript(ctx context.Context, id int) (*entity.Script, error) {
	var script entity.Script

	if err := r.queryRowxContextWithStructScan(
		ctx,
		fmt.Sprintf(`SELECT * FROM script WHERE id = %v`, id),
		&script,
	); err != nil {
		return nil, err
	}

	return &script, nil
}

func (r *Repo) GetAllScripts(ctx context.Context, offset, limit int) ([]*entity.Script, error) {
	query := "SELECT * FROM script ORDER BY created_at"

	if limit == math.MaxInt64 {
		query = fmt.Sprintf(`%v OFFSET %v`, query, offset)
	} else {
		query = fmt.Sprintf(`%v LIMIT %v OFFSET %v`, query, limit, offset)
	}

	rows, err := r.DB.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var scripts []*entity.Script

	if limit != math.MaxInt64 {
		scripts = make([]*entity.Script, 0, limit)
	}

	for rows.Next() {
		var script entity.Script

		if err = rows.StructScan(&script); err != nil {
			return nil, err
		}

		scripts = append(scripts, &script)
	}

	return scripts, nil
}
