package os

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// todo: https://stackoverflow.com/questions/69733417/golang-execute-shell-command-and-return-as-channel
func RunCommand(command string, out chan string) (int, error) { // todo: returning chan is bad practice?
	filename := fmt.Sprintf("./%v_temp_script.sh", time.Now().Unix())

	file, err := os.Create(filename)
	if err != nil {
		close(out)
		return -1, err
	}

	// close and then remove created file
	defer file.Close()
	defer os.Remove(filename)

	// todo: understand permissions
	if err = os.WriteFile(filename, []byte(command), 0666); err != nil {
		close(out)
		return -1, err
	}

	cmd := exec.Command("/bin/sh", filename)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		close(out)
		return -1, err
	}

	if err = cmd.Start(); err != nil {
		close(out)
		return -1, err
	}

	scanner := bufio.NewScanner(cmdReader)

	// todo: maybe use buffered chan?
	go func() {
		for scanner.Scan() {
			out <- scanner.Text()
		}
	}()

	pid := cmd.Process.Pid

	//if err = cmd.Wait(); err != nil {
	//	// return -1, err
	//	// TODO:
	//}

	go func() {
		if err = cmd.Wait(); err != nil { // TODO: i do not understand WHY in goroutine cmd.Wait() returns nil instantly, not waiting or running cmd?
			// return -1, err
			// TODO:
		}

		close(out)
	}()

	return pid, nil
}
