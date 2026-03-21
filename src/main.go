package main

import (
	"fmt"
	"time"
)

var currentProfile Profile
var loadedTools = map[string]Tool{}


func main() {
	ensureDirs()

	if !profileExists("default") {
		if !hasTTY() {
			fmt.Println("No TTY detected. Running in idle mode...")

			for {
				time.Sleep(10 * time.Second)
			}
		}
		must(runSetup())
	}

	must(loadProfile("default"))
	must(loadTools())

	
	
chatLoop()
	
}

