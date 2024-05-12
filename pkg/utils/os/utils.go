package os

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func RunCommand(ctx context.Context, command string, pidChan chan int, cmdChan chan *exec.Cmd, callback func(chan string)) error {
	errChan := make(chan error)
	doneChan := make(chan bool)
	outChan := make(chan string)

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

		cmdReader, err := cmd.StdoutPipe()
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

		scanner := bufio.NewScanner(cmdReader)
		for scanner.Scan() {
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
			return ErrContextCancelled

		case <-doneChan:
			return nil

		case err := <-errChan:
			return err
		}
	}
}
