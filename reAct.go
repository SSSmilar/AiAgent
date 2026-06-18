package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const reactSystem = `Ты решаешь задачу по циклу think → act → observe, без внешних инструментов.
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

	//Ограничил ReAct  5 запросами так как при долгом размышлении она может просто дизентигрировать токены .
	for i := 0; i < 5; i++ {
		//Вызываю запрос к LLM
		respMsg, err := ask(apiKey, Tools, reactSystem, dialogs)
		if err != nil {
			slog.Error("Error sending request to API ", "details", err)
			os.Exit(1)
		}
		//Оригинальный ответ модели добавляю 1 раз .
		dialogs = append(dialogs, respMsg)

		//Ветка с работы модели без вызова инструментов .
		if len(respMsg.ToolCalls) == 0 {
			//Вывожу  в консоль на каждом этапе чтобы видеть процесс размышления .
			fmt.Printf("\n=== Iteration %d ===\n ", i+1)
			fmt.Println(respMsg.Content)
			fmt.Println("=== End ===\n ")
			//Вывожу ответы в консоль чтобы видеть как модель думает .
			dialogs = append(dialogs, Message{Role: "assistant", Content: respMsg.Content})
			//Проверяю на финал , если не финал отправляю модель дальше думать .
			if strings.Contains(respMsg.Content, "FINAL") {
				break
			}
		}
		
		dialogs = append(dialogs, respMsg)
		for _, i := range respMsg.ToolCalls {
			if i.Function.Name == "GetContributors" {
				contributirs, err := GetContributors()
				if err != nil {
					slog.Error("Error receiving contributors ", "details", err)
					errors, err := json.Marshal(err)
					if err != nil {
						slog.Error("Error converting error to json ", "details", err)
					}
					dialogs = append(dialogs, Message{Role: "assistant", Content: string(errors)})
					continue
				}
				contributirsJSON, err := json.Marshal(contributirs)
				if err != nil {
					slog.Error("Error converting contributirs to json ", "details", err)
					errors, err := json.Marshal(err)
					if err != nil {
						slog.Error("Error converting error to json ", "details", err)
					}
					dialogs = append(dialogs, Message{Role: "assistant", Content: string(errors)})
					continue
				}
				dialogs = append(dialogs, Message{Role: "tool", Content: string(contributirsJSON), ToolCallID: i.ID})
				continue
			}
		}
		dialogs = append(dialogs, Message{Role: "user", Content: Observation})
	}
}
