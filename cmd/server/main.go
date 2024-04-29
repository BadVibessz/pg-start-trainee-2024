package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func splitInto(str string) (string, []string) {
	spltd := strings.Split(str, " ")

	return spltd[0], spltd[1:]
}

func main() {
	// name, args := splitInto("goose -dir db/migrations up")

	cmd := exec.Command("goose", "--help")

	var out strings.Builder
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf(out.String())
}
