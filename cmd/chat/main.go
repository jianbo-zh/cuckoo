package main

import (
	"github.com/jianbo-zh/dchat/cmd/chat/command"
	logging "github.com/jianbo-zh/go-log"
)

func main() {
	logging.SetupLogging(logging.Config{
		Level:  logging.LevelDebug,
		Format: logging.FormatPlaintextOutput,
		Stdout: true,
	})
	command.Execute()
}
