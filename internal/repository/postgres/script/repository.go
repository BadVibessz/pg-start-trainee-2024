package script

import (
	"context"
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
