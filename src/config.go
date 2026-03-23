package main

import (
	"os"
	"path/filepath"
)

var (
	cwd string
	baseDir    string
)

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	
	baseDir = filepath.Join(cwd, ".bear")
}

func personaDir(name string) string {
	return filepath.Join(baseDir, "personas", name)
}

func personaConfigPath(name string) string {
	return filepath.Join(personaDir(name), "config.json")
}

func personaToolsDir(name string) string {
	return filepath.Join(personaDir(name), "tools")
}

func personaChatsDir(name string) string {
	return filepath.Join(personaDir(name), "chats")
}

