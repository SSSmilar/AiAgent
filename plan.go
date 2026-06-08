package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const plannerSystem = "Ты - планировщик. Разбей задачу на 3–5 атомарных пунктов, по одному в строке. Только план, не выполняй."
const writerSystem = "Ты - исполнитель. Напиши ОДИН пункт кратко (2–3 предложения), не повторяя уже написанное."

func Plan(apiKey string, task string) {

	response, err := ask(apiKey, plannerSystem, []Message{{Role: "user", Content: task}})
	if err != nil {
		slog.Error("Error sending request to API ", "details", err)
		os.Exit(1)
	}
	//Рубим на пукты для работы .
	steps := strings.Split(response, "\n")
	//Массив для ответов .
	var done []string
	//Прохожусь по всем шагам и добовляю ответы в массив .
	for _, step := range steps {

		step = strings.TrimSpace(step)
		if step == "" {
			continue
		}
		alreadyWritten := strings.Join(done, "\n")

		prompt := fmt.Sprintf("Задача: %s\nПункт: %s\nУже написано:\n%s", task, step, alreadyWritten)

		executor := Message{Role: "user", Content: prompt}

		response, err := ask(apiKey, writerSystem, []Message{executor})
		if err != nil {
			slog.Error("Error sending request to API ", "details", err)
			os.Exit(1)
		}
		done = append(done, response)
	}
	//Вывожу ответы .
	slog.Info("Done", "details", done)
	for _, step := range done {
		fmt.Println(step)
	}
}
