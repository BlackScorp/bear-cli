package main

import (
	"encoding/json"
	"net/http"
	"bytes"
	"io"
)

type OllamaResp struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}



func askLLM(prompt string) (string, error) {
	body := map[string]interface{}{
		"model": currentProfile.Model,
		"stream": false,
		"messages": []map[string]string{
			{"role": "system", "content": buildSystemPrompt()},
			{"role": "user", "content": prompt},
		},
	}

	b, _ := json.Marshal(body)

	resp, err := http.Post(currentProfile.LLMURL+"/api/chat", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	var parsed OllamaResp
	if err := json.Unmarshal(data, &parsed); err != nil {
		return string(data), nil // fallback
	}

	return parsed.Message.Content, nil
}

func fetchModels(url string) ([]string, error) {
	resp, err := http.Get(url + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var parsed struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	b, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &parsed); err != nil {
		return nil, err
	}

	var names []string
	for _, m := range parsed.Models {
		names = append(names, m.Name)
	}

	return names, nil
}
