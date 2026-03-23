package main

import (
	"fmt"
	"time"
)

var currentProfile Profile

func main() {
	if !hasTTY() {
		fmt.Println("No TTY detected. Running in idle mode...")
		for {
			time.Sleep(10 * time.Second)
		}
	}

	if !profileExists("default") {
		must(runSetup())
	}

	must(loadProfile("default"))

	_, err := loadTools()
	must(err)

	chatLoop()
}