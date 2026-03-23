package main

import (
	"os"
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
	_, err := os.Stat(personaConfigPath(name))
	return err == nil
}

func saveProfile(p Profile) error {
	err := os.MkdirAll(personaDir(p.Name), 0755)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(personaConfigPath(p.Name), data, 0644)
}

func loadProfile(name string) error {
	data, err := os.ReadFile(personaConfigPath(name))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &currentProfile)
}