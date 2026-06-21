package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

// Message описывает одну реплику в диалоге.
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ChatRequest  описывает то что мы отправляем (POST body).
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
}

// ChatResponse описывает то что мы получаем в ответ.
type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

// Choice - это варант ответа от модели (по дефолту берём первый но можем ставить другой от модельки скейл обычно.
type Choice struct {
	Message Message `json:"message"`
}

// Contributor - структура которую я буду мапить в таблицу .
type Contributor struct {
	Login       string `json:"login"`
	CommitCount int    `json:"commit_count"`
}
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}
type ToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// ToolCall and FunctionCall Структуры для парсинга ответа модели .
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func GetContributors() (contributors []Contributor, err error) {
	databaseURL, err := GetDataBaseURL()
	if err != nil {
		slog.Error("Error receiving API KEY ", "details", err)
		return nil, err
	}
	ctx, ctxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer ctxCancel()
	dataBaseConnect, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		slog.Error("Error connecting to database ", "details", err)
		return nil, err
	}
	defer func() {
		if err := dataBaseConnect.Close(ctx); err != nil {
			slog.Error("Error closing database connection ", "details", err)
		}
	}()
	var contributorsData []Contributor
	rows, err := dataBaseConnect.Query(ctx, "SELECT login , commit_count FROM gitRepo")
	if err != nil {
		slog.Error("Error scanning contributors from database ", "details", err)
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		contributor := Contributor{}
		err := rows.Scan(&contributor.Login, &contributor.CommitCount)
		if err != nil {
			slog.Error("Error scanning contributors from database ", "details", err)
			continue
		}
		contributorsData = append(contributorsData, contributor)
	}
	if rows.Err() != nil {
		slog.Error("Error scanning contributors from database ", "details", rows.Err())
		return nil, rows.Err()
	}
	return contributorsData, nil
}

func GetDataBaseURL() (string, error) {

	err := godotenv.Load()
	if err != nil {
		return "", fmt.Errorf("error loading .env file: %w", err)
	}
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return "", fmt.Errorf("DATABASE_URL is not set")
	}
	return url, nil
}

func GetAPIKey() (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", fmt.Errorf("error loading .env file: %w", err)
	}
	apiKey := os.Getenv("GROQ_API_KEY")

	if apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY is not set")
	}
	return apiKey, nil
}

func main() {
	apiKey, err := GetAPIKey()
	if err != nil {
		slog.Error("Error receiving API KEY ", "details", err)
		os.Exit(1)
	}
	task := "Узнай, кто контрибьютил в наш репозиторий и выведи их логины"

	MyTools := []Tool{
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "GetContributors",
				Description: "Получает список контрибьюторов репозитория из базы данных ",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
	}
	Plan(apiKey, MyTools, task)
}
func ask(apiKey string, tool []Tool, system string, dialogs []Message) (Message, error) {
	messages := []Message{
		{Role: "system", Content: system},
	}
	messages = append(messages, dialogs...)
	reqBody := ChatRequest{
		Model:    "meta-llama/llama-4-scout-17b-16e-instruct", //Юзаю фри модель , но при сложных задачах можно будет просто сменить тут модель и пополнить счёт в Google AI Studio.
		Messages: messages,                                    //история диалога.
		Tools:    tool,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return Message{}, fmt.Errorf("failed to marshal request to JSON: %w", err)
	}
	//Google URL с поддержкой OpenAI.
	url := "https://api.groq.com/openai/v1/chat/completions"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return Message{}, fmt.Errorf("http request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	//Создаём клиента и делаем  запрос .
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("http response error: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("Error closing response body ", "details", err)
		}
	}()
	//Проверяем  статус ответа , так как если он не 200 мы получим бред после парсинга .
	if resp.StatusCode != http.StatusOK {
		errorBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return Message{}, fmt.Errorf("reading response error: %w", err)
		}
		return Message{}, fmt.Errorf("API error: status %d, details: %s", resp.StatusCode, string(errorBody))
	}
	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return Message{}, fmt.Errorf("decoding response error: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return Message{}, fmt.Errorf("model returned no choices")
	}
	return chatResp.Choices[0].Message, nil
}
