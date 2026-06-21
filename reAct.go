package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const reactSystem = `Ты решаешь задачу по циклу think → act → observe .
На каждом шаге выводи РОВНО один блок:

Thought: <короткое рассуждение>
Action: PROPOSE: <текущий вариант ответа>
   - или -
Action: FINAL: <ответ, в котором ты уверен>

После PROPOSE я пришлю Observation с просьбой перепроверить.
Перепроверь и либо исправь (снова PROPOSE), либо зафиксируй (FINAL).
Делай по одному шагу за раз, не выкладывай все решение сразу.`

const Observation = "Observation: перечитай свой вариант. Есть ошибка - исправь, иначе зафиксируй FINAL."

func ReAct(apiKey string, Tools []Tool, task string) {
	var dialogs []Message
	dialogs = append(dialogs, Message{Role: "user", Content: task})

	for i := 0; i < 5; i++ {
		respMsg, err := ask(apiKey, Tools, reactSystem, dialogs)
		if err != nil {
			slog.Error("Error sending request to API", "details", err)
			os.Exit(1)
		}

		dialogs = append(dialogs, respMsg)

		// ВЕТКА Б: Модель вызвала инструмент
		if len(respMsg.ToolCalls) > 0 {
			for _, toolCall := range respMsg.ToolCalls {
				if toolCall.Function.Name == "GetContributors" {
					contributors, err := GetContributors()
					var content string

					if err != nil {
						slog.Error("Database error", "details", err)
						content = fmt.Sprintf(`{"error": "%s"}`, err.Error())
					} else {
						contributorsJSON, _ := json.Marshal(contributors)
						content = string(contributorsJSON)
					}

					dialogs = append(dialogs, Message{
						Role:       "tool",
						Content:    content,
						ToolCallID: toolCall.ID,
					})
				}
			}
			continue
		}

		fmt.Printf("\n=== Iteration %d ===\n%s\n=== End ===\n", i+1, respMsg.Content)

		if strings.Contains(respMsg.Content, "FINAL") {
			break
		}
		
		dialogs = append(dialogs, Message{Role: "user", Content: Observation})
	}
}
