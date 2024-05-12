package script

import (
	"fmt"
	"pg-start-trainee-2024/domain/entity"
)

func getScriptSlice(n int) []entity.Script {
	res := make([]entity.Script, n)

	for i := 0; i < n; i++ {
		res[i] = entity.Script{
			Command:   fmt.Sprintf("command_%v", i+1),
			Output:    fmt.Sprintf("output of command_%v", i+1),
			IsRunning: true,
			PID:       i + 1*1000,
		}
	}

	return res
}

var (
	scripts = getScriptSlice(100)
)
