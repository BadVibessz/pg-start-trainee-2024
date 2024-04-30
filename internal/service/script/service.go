package script

import (
	"context"
	"os/exec"
	"pg-start-trainee-2024/domain/entity"
)

type Repo interface {
	CreateScript(ctx context.Context, script entity.Script) (*entity.Script, error)
}

type Service struct {
	Repo Repo
}

func New(repo Repo) *Service {
	return &Service{Repo: repo}
}

// CreateScript creates and runs new script
func (s *Service) CreateScript(ctx context.Context, script entity.Script) (*entity.Script, error) {
	scpt, err := s.Repo.CreateScript(ctx, script)
	if err != nil {
		return nil, err
	}

	exec.Command()
}
