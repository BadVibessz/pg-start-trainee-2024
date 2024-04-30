package main

import (
	"log"
	"strings"

	osutils "pg-start-trainee-2024/pkg/utils/os"
)

func splitInto(str string) (string, []string) {
	spltd := strings.Split(str, " ")

	return spltd[0], spltd[1:]
}

func main() {
	// name, args := splitInto("goose -dir db/migrations up")

	//args, err := shellwords.Parse("echo gunna | sudo -S systemctl status docker")
	//if err != nil {
	//	log.Fatal(err)
	//}

	//// todo: service layer must create new temp file with command inside it and then execute by exec.Command("/bin/sh", "filename")
	//cmd := exec.Command("/bin/sh", "./test_script.sh")
	//var out strings.Builder
	//cmd.Stdout = &out
	//
	//if err := cmd.Run(); err != nil {
	//	log.Fatal(err)
	//}

	//cmd := exec.Command("/bin/sh", "./test_script.sh")
	//cmd.Start()
	//cmd.Wait()

	output := make(chan string)

	pid, err := osutils.RunCommand("ping google.com", output)
	if err != nil {
		log.Fatalf(err.Error())
	}

	println(pid)

	for token := range output {
		print(token)
	}
}
