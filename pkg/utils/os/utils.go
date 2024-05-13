package os

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync/atomic"
	"time"
)

func RunCommand(ctx context.Context, command string, pidChan chan int, cmdChan chan *exec.Cmd, callback func(chan string)) error {
	errChan := make(chan error)
	doneChan := make(chan bool, 1)
	outChan := make(chan string)

	var isOutChanClosed int32

	defer close(outChan)

	handleErr := func(err error) {
		errChan <- err

		close(pidChan)
		close(cmdChan)
	}

	filename := fmt.Sprintf("./%v_temp_script.sh", time.Now().Unix())

	// create temp file
	if err := os.WriteFile(filename, []byte(command), 0666); err != nil {
		handleErr(err)
	}

	// remove created file
	defer os.Remove(filename)

	go func() {
		go func() {
			callback(outChan)
		}()

		cmd := exec.CommandContext(ctx, "/bin/sh", filename)

		stdoutReader, err := cmd.StdoutPipe()
		if err != nil {
			handleErr(err)
		}

		stderrReader, err := cmd.StderrPipe()
		if err != nil {
			handleErr(err)
		}

		if err = cmd.Start(); err != nil {
			handleErr(err)
		}

		pidChan <- cmd.Process.Pid
		close(pidChan)

		cmdChan <- cmd
		close(cmdChan)

		scanner := bufio.NewScanner(stdoutReader)
		for scanner.Scan() {
			if isOutChanClosed == 1 {
				break
			}

			outChan <- scanner.Text()
		}

		scanner = bufio.NewScanner(stderrReader)
		for scanner.Scan() {
			if isOutChanClosed == 1 {
				break
			}

			outChan <- scanner.Text()
		}

		if err = cmd.Wait(); err != nil {
			errChan <- err
		}

		doneChan <- true
	}()

	for {
		select {
		case <-ctx.Done():
			atomic.AddInt32(&isOutChanClosed, 1)

			return ErrContextCancelled

		case <-doneChan:
			atomic.AddInt32(&isOutChanClosed, 1)

			return nil

		case err := <-errChan:
			atomic.AddInt32(&isOutChanClosed, 1)

			return err
		}
	}
}
