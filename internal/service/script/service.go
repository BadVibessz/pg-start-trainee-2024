package script

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"pg-start-trainee-2024/domain/entity"

	osutils "pg-start-trainee-2024/pkg/utils/os"
)

type Repo interface {
	CreateScript(ctx context.Context, script entity.Script) (*entity.Script, error)
	UpdateScriptOutput(ctx context.Context, id int, output string) (*entity.Script, error)
	DeleteScript(ctx context.Context, id int) (*entity.Script, error)
	UpdateScriptPIDAndRunningState(ctx context.Context, id, pid int, isRunning bool) (*entity.Script, error)
	UpdateScriptRunningState(ctx context.Context, id int, isRunning bool) (*entity.Script, error)
	GetScript(ctx context.Context, id int) (*entity.Script, error)
	GetAllScripts(ctx context.Context, offset, limit int) ([]*entity.Script, error)
}

type Cache interface {
	Set(key string, value any, duration time.Duration)
	Get(key string) (any, bool)
	Delete(key string)
}

type Service struct {
	Repo Repo

	cacheMutex *sync.RWMutex
	Cache      Cache

	logger             *logrus.Logger
	outputBufferLength int
}

func New(repo Repo, cache Cache, outputBufferLength int) *Service {
	return &Service{
		Repo:               repo,
		cacheMutex:         &sync.RWMutex{},
		Cache:              cache,
		logger:             logrus.New(),
		outputBufferLength: outputBufferLength,
	}
}

func (s *Service) updateScriptOutputWithStrings(ctx context.Context, id int, strings ...string) (*entity.Script, error) {
	toAppend := ""
	for _, ss := range strings {
		toAppend += ss
	}

	return s.Repo.UpdateScriptOutput(ctx, id, toAppend)
}

func (s *Service) outCallback(ctx context.Context, n int, id int) func(chan string) {
	return func(outChan chan string) {
		strs := make([]string, n)

		i := 0
		for str := range outChan {
			if i < n {
				strs[i] = fmt.Sprintf("%v\n", str)
				i++
			} else {
				// update script and 'clear' strs
				i = 0

				_, err := s.updateScriptOutputWithStrings(ctx, id, strs...)
				if err != nil {
					s.logger.Errorf("error occurred udating script's output: %v", err)

					return
				}
			}
		}

		// chan is closed => update script with stored output in strs
		_, err := s.updateScriptOutputWithStrings(ctx, id, strs...)
		if err != nil {
			s.logger.Errorf("error occurred udating script's output: %v", err)

			return
		}
	}
}

// CreateScript creates and runs new script
func (s *Service) CreateScript(ctx context.Context, script entity.Script) (*entity.Script, error) {
	pidChan := make(chan int, 1)
	cmdChan := make(chan *exec.Cmd, 1)

	scpt, err := s.Repo.CreateScript(ctx, script)
	if err != nil {
		return nil, err
	}

	scptMutex := &sync.RWMutex{}

	cmdCtx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			if pid := <-pidChan; pid != 0 {
				// just as we captured pid => we can update script's PID
				var updateErr error

				scptMutex.Lock()

				scpt, updateErr = s.Repo.UpdateScriptPIDAndRunningState(ctx, scpt.ID, pid, true)
				if updateErr != nil {
					s.logger.Errorf("error occurred updating script's PID and running state: %v", err)
				}

				scptMutex.Unlock()

				break
			}
		}
	}()

	var cmd *exec.Cmd

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			if c := <-cmdChan; c != nil {
				cmd = c

				break
			}
		}
	}()

	go func() {
		runErr := osutils.RunCommand(
			cmdCtx,
			script.Command,
			pidChan,
			cmdChan,
			s.outCallback(cmdCtx, s.outputBufferLength, scpt.ID),
		)

		if runErr != nil {
			s.logger.Errorf("error occurred running script: %v", runErr)
		}

		// script execution not started or stopped: whether it's cancelled or not -> update is_running to false
		scptMutex.RLock()
		if _, updateErr := s.Repo.UpdateScriptRunningState(context.Background(), scpt.ID, false); updateErr != nil {
			s.logger.Errorf("error occurred updating script is_running column: %v", err)
		}
		scptMutex.RUnlock()

		// remove from cache
		s.cacheMutex.Lock()
		defer s.cacheMutex.Unlock()

		s.Cache.Delete(strconv.Itoa(scpt.ID))
	}()

	wg.Wait()

	// as script started we can add to inmemory cache tuple (script.ID, context.CancelFunc)
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.Cache.Set(strconv.Itoa(scpt.ID), entity.CmdContext{Cmd: cmd, Cancel: cancel}, -1)

	return scpt, err
}

func (s *Service) StopScript(ctx context.Context, id int) error {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	cmdContextAny, exist := s.Cache.Get(strconv.Itoa(id))
	if !exist {
		return ErrNoSuchRunningScript
	}

	cmdContext, ok := cmdContextAny.(entity.CmdContext)
	if !ok {
		return ErrCannotCastToCancelFunc
	}

	cmdContext.Cancel()

	if _, err := s.Repo.UpdateScriptRunningState(ctx, id, false); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetAllScripts(ctx context.Context, offset, limit int) ([]*entity.Script, error) {
	return s.Repo.GetAllScripts(ctx, offset, limit)
}

func (s *Service) GetScript(ctx context.Context, id int) (*entity.Script, error) {
	return s.Repo.GetScript(ctx, id)
}

func (s *Service) DeleteScript(ctx context.Context, id int) error {
	_, err := s.GetScript(ctx, id)
	if err != nil {
		return ErrNoSuchScript
	}

	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	if cmdContextAny, exist := s.Cache.Get(strconv.Itoa(id)); exist {
		cmdContext, ok := cmdContextAny.(entity.CmdContext)
		if ok {
			cmdContext.Cancel()
		}
	}

	if _, err = s.Repo.DeleteScript(ctx, id); err != nil {
		return err
	}

	return nil
}
