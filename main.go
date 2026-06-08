package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

// Message описывает одну реплику в диалоге.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest  описывает то что мы отправляем (POST body).
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// ChatResponse описывает то что мы получаем в ответ.
type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

// Choice - это варант ответа от модели (по дефолту берём первый но можем ставить другой от модельки скейл обычно.
type Choice struct {
	Message Message `json:"message"`
}

func GetApiKey() (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", fmt.Errorf("error loading .env file: %w", err)
	}
	apiKey := os.Getenv("OPENAI_API_KEY")

	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY is not set")
	}
	return apiKey, nil
}
func main() {
	apiKey, err := GetApiKey()
	if err != nil {
		slog.Error("Ошибка получения API KEY", err)
		os.Exit(1)
	}
	
}
