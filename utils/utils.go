package utils

import (

	"log"
	"os"
	"resty.dev/v3"
)


func ResponseTextFromAi(text string) string {
	// Add your text simplification logic here

	prompt := `
      You are a helpful assistant. 
      The following is an email body. 
      Please rewrite this in simple, layman terms so a non-expert can understand it immediately.
      Structure it with bullet points. 
	  If there are any financial terms or technical jargon, please explain them in simple terms.
	  If the email contains action items, please highlight them clearly.
	  If the email contains links, please list them at the end under "Relevant Links".
	  Email Body:
	  ` + text + `
	`

	response := CallAiAPI(prompt)
	return response
}

type AiRequest struct {
	Model    string        `json:"model"`
	Messages []AiMessage  `json:"messages"`
}

type AiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AiResponse struct {
	Choices []struct {
		Message AiMessage `json:"message"`
	} `json:"choices"`
}	

func CallAiAPI(prompt string) string {
	url := os.Getenv("AI_URL")
	apiKey := os.Getenv("AI_API_KEY")
	model := os.Getenv("AI_MODEL")
	client := resty.New()
	defer client.Close()
	
	var aiRequest AiRequest
	aiRequest.Model = model
	aiRequest.Messages = []AiMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+apiKey).
		SetBody(aiRequest).
		SetResult(&AiResponse{}).
		Post(url)

	if err != nil {
		log.Fatalf("Error making API request: %v", err)
	}

	aiResponse := res.Result().(*AiResponse)
	if len(aiResponse.Choices) > 0 {
		return aiResponse.Choices[0].Message.Content
	} else {
		log.Println("No choices returned in AI response")
		return ""
	}
}