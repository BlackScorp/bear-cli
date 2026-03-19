package main
import (
	"os"
	"path/filepath"
	"strings"
	"encoding/json"
	"errors"
	"time"
	"context"
	"os/exec"
)

type Tool struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Exec        string   `json:"exec"`
	Arguments   []string `json:"arguments"`
	File        string   `json:"-"` // internal: filename source
}

type ToolCall struct {
	Tool      string            `json:"tool"`
	Arguments map[string]string `json:"arguments"`
}

func loadTools() error {
	files, err := os.ReadDir(toolsDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(toolsDir, f.Name()))
		if err != nil {
			continue
		}

		var t Tool
		if err := json.Unmarshal(data, &t); err != nil {
			continue
		}

		loadedTools[t.Name] = t
	}
	return nil
}

func runTool(name string, args map[string]string) (string, error) {
	tool, ok := loadedTools[name]
	if !ok {
		return "", errors.New("tool not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var cmdArgs []string
	for _, a := range tool.Arguments {
		cmdArgs = append(cmdArgs, args[a])
	}

	cmd := exec.CommandContext(ctx, tool.Exec, cmdArgs...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func tryParseTool(resp string) (ToolCall, bool) {
	var tc ToolCall
	if json.Unmarshal([]byte(resp), &tc) == nil && tc.Tool != "" {
		return tc, true
	}
	return tc, false
}
