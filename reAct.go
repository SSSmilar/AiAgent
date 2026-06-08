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

func ReAct(apiKey string, task string) {
	dialogs := []Message{{
		Role:    "user",
		Content: task,
	}}
	//Ограничил ReAct  5 запросами так как при долгом размышлении она может просто дизентигрировать токены .
	for i := 0; i < 5; i++ {
		//Вызываю запрос к LLM
		response, err := ask(apiKey, reactSystem, dialogs)
		if err != nil {
			slog.Error("Error sending request to API ", "details", err)
			os.Exit(1)
		}
		//Вывожу  в консоль на каждом этапе чтобы видеть процесс размышления .
		fmt.Printf("\n=== Iteration %d ===\n ", i+1)
		fmt.Println(response)
		fmt.Println("=== End ===\n ")
		//Вывожу ответы в консоль чтобы видеть как модель думает .
		dialogs = append(dialogs, Message{Role: "assistant", Content: response})
		//Проверяю на финал , если не финал отправляю модель дальше думать .
		if strings.Contains(response, "FINAL") {
			break
		}

		dialogs = append(dialogs, Message{Role: "user", Content: Observation})
	}
}
