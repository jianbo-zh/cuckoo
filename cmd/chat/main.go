package main

import (
	"github.com/jianbo-zh/dchat/cmd/chat/command"
	logging "github.com/jianbo-zh/go-log"
)

func main() {
	logging.SetLogLevel("*", "debug")
	command.Execute()
}
