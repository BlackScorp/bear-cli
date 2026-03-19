package main

import (
	"os"
	"path/filepath"
	"encoding/json"
)
type Profile struct {
	Name       string   `json:"name"`
	LLMURL     string   `json:"llm_url"`
	Model      string   `json:"model"`
	Tools      []string `json:"tools"`
	BasePrompt string   `json:"base_prompt"`
}


func profileExists(name string) bool {
	_, err := os.Stat(filepath.Join(profileDir, name+".json"))
	return err == nil
}

func saveProfile(p Profile) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(profileDir, p.Name+".json"), data, 0644)
}

func loadProfile(name string) error {
	data, err := os.ReadFile(filepath.Join(profileDir, name+".json"))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &currentProfile)
}