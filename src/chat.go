package main

import (
	"bufio"
	"os"
	"fmt"
	"strings"
	"path/filepath"
	"time"
)


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