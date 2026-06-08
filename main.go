package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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
		slog.Error("Error receiving API KEY ", "details", err)
		os.Exit(1)
	}
	task := "У фермера 17 овец. Все, кроме 9, убежали. Сколько осталось?"

	ReAct(apiKey, task)
}
func ask(apiKey string, system string, dialogs []Message) (string, error) {
	messages := []Message{
		{Role: "system", Content: system},
	}
	messages = append(messages, dialogs...)
	reqBody := ChatRequest{
		Model:    "gemini-2.5-flash", //Юзаю фри модель , но при сложных задачах можно будет просто сменить тут модель и пополнить счёт в Google AI Studio.
		Messages: messages,           //история диалога.
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request to JSON: %w", err)
	}
	//Google URL с поддержкой OpenAI.
	url := "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("Err HTTP request: %w ", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	//Создаём клиента и делаем  запрос .
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Err HTTP response: %w ", err)
	}
	defer resp.Body.Close()
	//Проверяем  статус ответа , так как если он не 200 мы получим бред после парсинга .
	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: status %d, details: %s", resp.StatusCode, string(errorBody))
	}
	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("Err decoding response: %w ", err)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("model returned no choices ")
	}
	return chatResp.Choices[0].Message.Content, nil
}
