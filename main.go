package main

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

func main() {

}
