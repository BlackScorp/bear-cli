package main

import (
	"os"
	"fmt"
	"strings"
	"bufio"
)

func runSetup() error {
	s := bufio.NewScanner(os.Stdin)

	fmt.Print("LLM URL (http://localhost:11434): ")
	
	url, err := readLine(s)
	if err != nil {
		return fmt.Errorf("setup requires interactive input")
	}

	if url == "" {
		url = "http://127.0.0.1:11434"
	}

	models, err := fetchModels(url)
	if err != nil || len(models) == 0 {
		return fmt.Errorf("failed to fetch models: %w", err)
	}

	fmt.Println("Available models:")
	for i, m := range models {
		fmt.Printf("%d: %s\n", i, m)
	}

	fmt.Print("Select model index: ")
	s.Scan()
	var idx int
	fmt.Sscanf(s.Text(), "%d", &idx)
	if idx < 0 || idx >= len(models) {
		idx = 0
	}

	fmt.Print("Profile name (default): ")
	s.Scan()
	name := strings.TrimSpace(s.Text())
	if name == "" {
		name = "default"
	}

	fmt.Print("Base prompt (optional): ")
	s.Scan()
	basePrompt := s.Text()

	p := Profile{
		Name:       name,
		LLMURL:  url,
		Model:      models[idx],
		Tools:      []string{},
		BasePrompt: basePrompt,
	}

	return saveProfile(p)
}
func readLine(s *bufio.Scanner) (string, error) {
	if !s.Scan() {
		return "", fmt.Errorf("no input available")
	}
	return strings.TrimSpace(s.Text()), nil
}

func ensureDirs() {
	os.MkdirAll(profileDir, 0755)
	os.MkdirAll(toolsDir, 0755)
	os.MkdirAll(chatsDir, 0755)
}

func must(err error) {
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
}
func hasTTY() bool {
	return isTerminal(os.Stdin) && isTerminal(os.Stdout)
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
func isStdinAvailable() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	
	return (fi.Mode() & os.ModeCharDevice) != 0
}