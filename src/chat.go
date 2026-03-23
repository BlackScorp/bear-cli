package main

import (
	"bufio"
	"os"
	"fmt"
	"strings"
	"path/filepath"
	"time"
	"text/template"
	"embed"
)

//go:embed prompts/*.md
var promptFS embed.FS

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

		for {
			cmd, ok := extractCommand(resp)
			if !ok {
				break 
			}

			result, err := runCommand(cmd)
			if err != nil {
				result = err.Error()
			}
			result = limitOutput(result)

		
			resp, err = askLLM("Command output:\n" + result)
			if err != nil {
				fmt.Println("error:", err)
				break
			}
		}

		fmt.Println(resp)
		saveChat(input, resp)
	}
}

func saveChat(input, resp string) {
	file := filepath.Join(personaChatsDir(currentProfile.Name), time.Now().Format("2006-01-02")+".md")

	f, _ := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	f.WriteString(fmt.Sprintf("\n**SystemPrompt**: %s\n\n**User:** %s\n\n**AI:** %s\n",  buildSystemPrompt(), input, resp))
}

func showHistory() {
	files, _ := os.ReadDir(personaChatsDir(currentProfile.Name))
	for _, f := range files {
		fmt.Println(f.Name())
	}
}

func loadSystemPromptTemplate() (string, error) {
	data, err := promptFS.ReadFile("prompts/system.md")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func buildSystemPrompt() string {
	
	tmplStr, err := loadSystemPromptTemplate()
	if err != nil {
		return "failed to load system prompt"
	}

	tmpl, err := template.New("prompt").Parse(string(tmplStr))
	if err != nil {
		return "failed to parse template"
	}

	cwd, _ := os.Getwd()

	data := map[string]interface{}{
		"BasePrompt": currentProfile.BasePrompt,
		"CWD":        cwd,
		"Tools":      buildToolsPrompt(),
	}

	var out strings.Builder
	err = tmpl.Execute(&out, data)
	if err != nil {
		return "failed to render prompt"
	}

	return out.String()
}


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

func limitOutput(s string) string {
	if len(s) > 2000 {
		return s[:2000] + "\n...[truncated]"
	}
	return s
}