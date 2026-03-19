package main

import (
	"os"
	"path/filepath"
)

var (
	baseDir    = filepath.Join(os.Getenv("HOME"), ".baer")
	profileDir = filepath.Join(baseDir, "profiles")
	toolsDir   = filepath.Join(baseDir, "tools")
	chatsDir   = filepath.Join(baseDir, "chats")
)