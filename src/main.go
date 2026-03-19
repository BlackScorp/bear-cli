package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ---------------- CONFIG ----------------

var (
	baseDir    = filepath.Join(os.Getenv("HOME"), ".baer")
	profileDir = filepath.Join(baseDir, "profiles")
	toolsDir   = filepath.Join(baseDir, "tools")
	chatsDir   = filepath.Join(baseDir, "chats")
)

// ---------------- DATA ----------------

type Profile struct {
	Name       string   `json:"name"`
	LLMURL     string   `json:"llm_url"`
	Model      string   `json:"model"`
	Tools      []string `json:"tools"`
	BasePrompt string   `json:"base_prompt"`
}

type Tool struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Exec        string   `json:"exec"`
	Arguments   []string `json:"arguments"`
	File        string   `json:"-"` // internal: filename source
}

// Ollama response (minimal robust)
type OllamaResp struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

// Tool call schema
type ToolCall struct {
	Tool      string            `json:"tool"`
	Arguments map[string]string `json:"arguments"`
}

// ---------------- GLOBAL STATE ----------------

var currentProfile Profile
var loadedTools = map[string]Tool{}

// ---------------- MAIN ----------------

func main() {
	ensureDirs()

	if !profileExists("default") {
		must(runSetup())
	}

	must(loadProfile("default"))
	must(loadTools())

	chatLoop()
}

// ---------------- SETUP ----------------

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

// ---------------- PROFILE ----------------

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

// ---------------- TOOLS ----------------

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

// ---------------- CHAT ----------------

func chatLoop() {
	s := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		s.Scan()
		input := s.Text()

		if strings.HasPrefix(input, "/") {
			handleCommand(input)
			continue
		}

		resp, err := askLLM(input)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}

		// Try tool call
		if tc, ok := tryParseTool(resp); ok {
			result, err := runTool(tc.Tool, tc.Arguments)
			if err != nil {
				result = err.Error()
			}

			resp, _ = askLLM("Tool result:\n" + result)
		}

		fmt.Println(resp)
		saveChat(input, resp)
	}
}

// ---------------- COMMANDS ----------------

func handleCommand(cmd string) {
	parts := strings.Split(cmd, " ")

	switch parts[0] {
	case "/exit":
		os.Exit(0)
	case "/history":
		showHistory()
	default:
		fmt.Println("unknown command")
	}
}

// ---------------- OLLAMA ----------------

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

// ---------------- TOOL PARSING ----------------

func tryParseTool(resp string) (ToolCall, bool) {
	var tc ToolCall
	if json.Unmarshal([]byte(resp), &tc) == nil && tc.Tool != "" {
		return tc, true
	}
	return tc, false
}

func buildSystemPrompt() string {
	var sb strings.Builder

	sb.WriteString(currentProfile.BasePrompt + "\n\n")
	sb.WriteString("You can use tools.\n")

	for _, t := range loadedTools {
		sb.WriteString(fmt.Sprintf("- %s(%v): %s\n", t.Name, t.Arguments, t.Description))
	}

	sb.WriteString("\nReturn ONLY JSON when calling tools.\n")

	return sb.String()
}

// ---------------- HISTORY ----------------

func saveChat(input, resp string) {
	file := filepath.Join(chatsDir, time.Now().Format("2006-01-02")+".md")

	f, _ := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	f.WriteString(fmt.Sprintf("\n**User:** %s\n\n**AI:** %s\n", input, resp))
}

func showHistory() {
	files, _ := os.ReadDir(chatsDir)
	for _, f := range files {
		fmt.Println(f.Name())
	}
}

// ---------------- UTILS ----------------

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
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ---------------- CONFIG ----------------

var (
	baseDir    = filepath.Join(os.Getenv("HOME"), ".baer")
	profileDir = filepath.Join(baseDir, "profiles")
	toolsDir   = filepath.Join(baseDir, "tools")
	chatsDir   = filepath.Join(baseDir, "chats")
)

// ---------------- DATA ----------------

type Profile struct {
	Name       string   `json:"name"`
	LLMURL     string   `json:"llm_url"`
	Model      string   `json:"model"`
	Tools      []string `json:"tools"`
	BasePrompt string   `json:"base_prompt"`
}

type Tool struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Exec        string   `json:"exec"`
	Arguments   []string `json:"arguments"`
	File        string   `json:"-"` // internal: filename source
}

// Ollama response (minimal robust)
type OllamaResp struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

// Tool call schema
type ToolCall struct {
	Tool      string            `json:"tool"`
	Arguments map[string]string `json:"arguments"`
}

// ---------------- GLOBAL STATE ----------------

var currentProfile Profile
var loadedTools = map[string]Tool{}

// ---------------- MAIN ----------------

func main() {
	ensureDirs()

	if !profileExists("default") {
		must(runSetup())
	}

	must(loadProfile("default"))
	must(loadTools())

	chatLoop()
}

// ---------------- SETUP ----------------

func runSetup() error {
	s := bufio.NewScanner(os.Stdin)

	fmt.Print("LLM URL (http://localhost:11434): ")
	s.Scan()
	url := strings.TrimSpace(s.Text())
	if url == "" {
		url = "http://localhost:11434"
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

// ---------------- PROFILE ----------------

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

// ---------------- TOOLS ----------------

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

// ---------------- CHAT ----------------

func chatLoop() {
	s := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		s.Scan()
		input := s.Text()

		if strings.HasPrefix(input, "/") {
			handleCommand(input)
			continue
		}

		resp, err := askLLM(input)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}

		// Try tool call
		if tc, ok := tryParseTool(resp); ok {
			result, err := runTool(tc.Tool, tc.Arguments)
			if err != nil {
				result = err.Error()
			}

			resp, _ = askLLM("Tool result:\n" + result)
		}

		fmt.Println(resp)
		saveChat(input, resp)
	}
}

// ---------------- COMMANDS ----------------

func handleCommand(cmd string) {
	parts := strings.Split(cmd, " ")

	switch parts[0] {
	case "/exit":
		os.Exit(0)
	case "/history":
		showHistory()
	default:
		fmt.Println("unknown command")
	}
}

// ---------------- OLLAMA ----------------

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

// ---------------- TOOL PARSING ----------------

func tryParseTool(resp string) (ToolCall, bool) {
	var tc ToolCall
	if json.Unmarshal([]byte(resp), &tc) == nil && tc.Tool != "" {
		return tc, true
	}
	return tc, false
}

func buildSystemPrompt() string {
	var sb strings.Builder

	sb.WriteString(currentProfile.BasePrompt + "\n\n")
	sb.WriteString("You can use tools.\n")

	for _, t := range loadedTools {
		sb.WriteString(fmt.Sprintf("- %s(%v): %s\n", t.Name, t.Arguments, t.Description))
	}

	sb.WriteString("\nReturn ONLY JSON when calling tools.\n")

	return sb.String()
}

// ---------------- HISTORY ----------------

func saveChat(input, resp string) {
	file := filepath.Join(chatsDir, time.Now().Format("2006-01-02")+".md")

	f, _ := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	f.WriteString(fmt.Sprintf("\n**User:** %s\n\n**AI:** %s\n", input, resp))
}

func showHistory() {
	files, _ := os.ReadDir(chatsDir)
	for _, f := range files {
		fmt.Println(f.Name())
	}
}

// ---------------- UTILS ----------------

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
