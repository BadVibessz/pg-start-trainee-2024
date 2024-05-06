package os

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// todo: https://stackoverflow.com/questions/69733417/golang-execute-shell-command-and-return-as-channel
func RunCommand(ctx context.Context, command string, pidChan chan int, callback func(chan string)) error {
	errChan := make(chan error)
	doneChan := make(chan bool)
	outChan := make(chan string)

	filename := fmt.Sprintf("./%v_temp_script.sh", time.Now().Unix())

	// create temp file
	if err := os.WriteFile(filename, []byte(command), 0666); err != nil {
		errChan <- err
	}

	// remove created file
	defer os.Remove(filename)

	// defer close(outChan)
	// defer close(pidChan) // todo?

	go func() {
		go func() {
			callback(outChan)
		}()

		cmd := exec.Command("/bin/sh", filename)

		cmdReader, err := cmd.StdoutPipe()
		if err != nil {
			errChan <- err
		}

		if err = cmd.Start(); err != nil {
			errChan <- err
		}

		pidChan <- cmd.Process.Pid
		close(pidChan)

		// todo: maybe use buffered chan?
		// todo: go func here?
		scanner := bufio.NewScanner(cmdReader)
		for scanner.Scan() {
			outChan <- scanner.Text()
		}

		// close(outChan)

		if err = cmd.Wait(); err != nil {
			//return err
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

//func Run(command string, callback func(chan string)) (int, chan error){
//
//}
