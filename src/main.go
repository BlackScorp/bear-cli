package main

import (
	"fmt"
	"os"
	"strings"
)

var currentProfile Profile
var loadedTools = map[string]Tool{}


func main() {
	ensureDirs()

	if !profileExists("default") {
		must(runSetup())
	}

	must(loadProfile("default"))
	must(loadTools())

	chatLoop()
}



func handleCommand(cmd string) {
	parts := strings.Split(cmd, " ")

	switch parts[0] {
	case "/exit":
		os.Exit(0)
	case "/history":
		showHistory()
	default:
		fmt.Println("unknown command")
	}
}
