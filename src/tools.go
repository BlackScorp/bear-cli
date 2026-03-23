package main

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
	"time"
	"context"
	"os/exec"
	"log"
)



//go:embed tools/*.md
var toolTemplates embed.FS

func installTool(toolName, targetDir string) error {
	data, err := toolTemplates.ReadFile("tools/" + toolName)
	if err != nil {
		return err
	}
	path := filepath.Join(targetDir, toolName)
	return os.WriteFile(path, data, 0644)
}

func listEmbeddedTools() ([]string, error) {
	files, err := toolTemplates.ReadDir("tools")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, f := range files {
		names = append(names, f.Name())
	}
	return names, nil
}

func loadTools() ([]string, error) {
	dir := personaToolsDir(currentProfile.Name)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var tools []string
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			continue
		}
		tools = append(tools, string(data))
	}
	return tools, nil
}

func buildToolsPrompt() string {
	tools, err := loadTools()
	if err != nil || len(tools) == 0 {
		return ""
	}
	return strings.Join(tools, "\n\n")
}

func extractCommand(resp string) (string, bool) {
	resp = strings.TrimSpace(resp)
	if strings.Contains(resp, "```bash") {
		start := strings.Index(resp, "```bash") + len("```bash")
		end := strings.Index(resp[start:], "```")
		if end == -1 {
			return "", false
		}
		cmd := resp[start : start+end]
		return strings.TrimSpace(cmd), true
	}
	return "", false
}

func runCommand(command string) (string, error) {
	log.Println("[CMD] Running:", command)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	out, err := cmd.CombinedOutput()
	log.Println("[CMD] Output:", string(out))
	return string(out), err
}
