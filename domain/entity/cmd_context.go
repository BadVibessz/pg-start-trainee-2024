package entity

import (
	"context"
	"os/exec"
)

// todo: in what package store this?

type CmdContext struct {
	Cancel context.CancelFunc
	Cmd    *exec.Cmd
}
