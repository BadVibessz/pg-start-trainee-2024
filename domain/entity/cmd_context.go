package entity

import (
	"context"
	"os/exec"
)

type CmdContext struct {
	Cancel context.CancelFunc
	Cmd    *exec.Cmd
}
