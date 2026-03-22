package main
import (
	"os"
	"path/filepath"
	"strings"
	"encoding/json"
	"time"
	"log"
	"fmt"
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
		log.Printf("[Tool] ❌ tool not found: %s\n", name)
		return "", fmt.Errorf("tool not found: %s", name)
	}

	// build args in defined order
	var cmdArgs []string
	for _, argName := range tool.Arguments {
		val, ok := args[argName]
		if !ok {
			errMsg := fmt.Sprintf("missing argument: %s", argName)
			log.Printf("[Tool] ❌ %s\n", errMsg)
			return "", fmt.Errorf(errMsg)
		}
		cmdArgs = append(cmdArgs, val)
	}

	// Debug: what will be executed
	log.Printf("[Tool] 🛠 Running: %s %v\n", tool.Exec, cmdArgs)

	// run with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, tool.Exec, cmdArgs...)
	out, err := cmd.CombinedOutput()

	// Debug: Raw tool output
	log.Printf("[Tool] 📤 Output: %s\n", string(out))

	if err != nil {
		// Debug: Error info
		log.Printf("[Tool] ❗ Error: %v\n", err)
		return string(out), err
	}

	log.Printf("[Tool] ✅ Success\n")

	return string(out), nil
}

func tryParseTool(resp string) (ToolCall, bool) {
	var tc ToolCall

	clean := cleanJSON(resp)

	if json.Unmarshal([]byte(clean), &tc) == nil && tc.Tool != "" {
		return tc, true
	}
	return tc, false
}

func cleanJSON(input string) string {
	input = strings.TrimSpace(input)

	// remove markdown fences if present
	input = strings.TrimPrefix(input, "```json")
	input = strings.TrimPrefix(input, "```")
	input = strings.TrimSuffix(input, "```")

	return strings.TrimSpace(input)
}