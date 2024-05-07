package script

import "C"
import (
	"context"
	"fmt"
	"pg-start-trainee-2024/domain/entity"
	"strconv"
	"sync"
	"time"

	osutils "pg-start-trainee-2024/pkg/utils/os"
)

type Repo interface {
	CreateScript(ctx context.Context, script entity.Script) (*entity.Script, error)
	UpdateScriptOutput(ctx context.Context, id int, output string) (*entity.Script, error)
	DeleteScript(ctx context.Context, id int) (*entity.Script, error)
	UpdateScriptPID(ctx context.Context, id, pid int) (*entity.Script, error)
}

type Cache interface {
	Set(key string, value any, duration time.Duration)
	Get(key string) (any, bool)
}

type Service struct {
	Repo  Repo
	Cache Cache
}

func New(repo Repo, cache Cache) *Service {
	return &Service{
		Repo:  repo,
		Cache: cache,
	}
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

	cmdCtx, cancel := context.WithCancel(context.Background())

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
			cmdCtx,
			script.Command,
			pidChan,
			s.outCallback(cmdCtx, 5, scpt.ID),
		); err != nil { // todo: delete only if err != ErrContextCancelled
			// if err occurred => we have to delete created script from db
			if _, err := s.Repo.DeleteScript(ctx, scpt.ID); err != nil {
				// todo:
			}
		}
	}()

	wg.Wait()

	// as script started we can add to inmemory cache tuple (script.ID, context.CancelFunc)
	s.Cache.Set(strconv.Itoa(scpt.ID), cancel, -1) // todo: duration forever? RAM overflow?

	return scpt, err
}

func (s *Service) StopScript(ctx context.Context, id int) error {
	cancelAny, exist := s.Cache.Get(strconv.Itoa(id))
	if !exist {
		return ErrNoSuchRunningScript
	}

	cancel, ok := cancelAny.(context.CancelFunc)
	if !ok {
		return ErrCannotCastToCancelFunc
	}

	cancel()

	// todo: update is_running state in db

	return nil
}
