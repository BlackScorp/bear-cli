package main

import (
	"os"
	"path/filepath"
)

var (
	baseDir    string
	profileDir string
	toolsDir   string
	chatsDir   string
)

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	baseDir = filepath.Join(cwd, ".bear")
	profileDir = filepath.Join(baseDir, "profiles")
	toolsDir = filepath.Join(baseDir, "tools")
	chatsDir = filepath.Join(baseDir, "chats")
}