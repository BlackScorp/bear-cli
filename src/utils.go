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
	s.Scan()
	url := strings.TrimSpace(s.Text())
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
