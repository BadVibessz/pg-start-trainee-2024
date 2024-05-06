package script

import (
	"context"
	"fmt"
	"pg-start-trainee-2024/domain/entity"
	"sync"

	osutils "pg-start-trainee-2024/pkg/utils/os"
)

type Repo interface {
	CreateScript(ctx context.Context, script entity.Script) (*entity.Script, error)
	UpdateScriptOutput(ctx context.Context, id int, output string) (*entity.Script, error)
	DeleteScript(ctx context.Context, id int) (*entity.Script, error)
	UpdateScriptPID(ctx context.Context, id, pid int) (*entity.Script, error)
}

type Service struct {
	Repo Repo
}

func New(repo Repo) *Service {
	return &Service{Repo: repo}
}

func (s *Service) updateScriptOutputWithStrings(ctx context.Context, id int, strings ...string) (*entity.Script, error) {
	toAppend := ""
	for _, ss := range strings {
		toAppend += ss
	}

	return s.Repo.UpdateScriptOutput(ctx, id, toAppend)
}

func (s *Service) outCallback(ctx context.Context, n int, id int) func(chan string) { // todo: understand how often to update script in db, n = 50? what n is optimal?
	return func(outChan chan string) {
		strs := make([]string, n)

		i := 0
		for str := range outChan {
			if i < n {
				strs = append(strs, fmt.Sprintf("%v\n", str))
				i++
			} else {
				// update script and 'clear' strs
				i = 0

				_, err := s.updateScriptOutputWithStrings(ctx, id, strs...)
				if err != nil {
					return // todo:
				}
			}
		}

		// chan is closed => update script with stored output in strs
		_, err := s.updateScriptOutputWithStrings(ctx, id, strs...)
		if err != nil {
			return // todo:
		}
	}
}

// CreateScript creates and runs new script
func (s *Service) CreateScript(ctx context.Context, script entity.Script) (*entity.Script, error) {
	pidChan := make(chan int, 1)

	scpt, err := s.Repo.CreateScript(ctx, script)
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			if pid := <-pidChan; pid != 0 {
				// just as we captured pid => we can update script's PID
				scpt, err = s.Repo.UpdateScriptPID(ctx, scpt.ID, pid) // todo: is_running := true

				break
			}
		}

	}()

	go func() {
		if err := osutils.RunCommand(
			ctx,
			script.Command,
			pidChan,
			s.outCallback(ctx, 5, scpt.ID),
		); err != nil { // todo: delete only if err != ErrContextCancelled
			// if err occurred => we have to delete created script from db
			if _, err := s.Repo.DeleteScript(ctx, scpt.ID); err != nil {
				// todo:
			}
		}
	}()

	wg.Wait()

	return scpt, err
}
