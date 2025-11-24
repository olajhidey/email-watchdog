package utils

import (
	"context"
	"log"
	"os"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
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
	Model    string      `json:"model"`
	Messages []AiMessage `json:"messages"`
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

type EmailSummary struct {
	TimeDate string `json:"time_date"`
	Summary  string `json:"summary"`
}

func getDateTime() string {
	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")
	return formattedTime
}

func StoreResponseToDatabase(firebase *firebase.App, response string) error {
	// Store response to firestore - Store current datetime and the response in the firestore db
	ctx := context.Background()
	client, err := firebase.Firestore(ctx)
	if err != nil {
		log.Fatalln("Error loading firestore")
	}
	defer client.Close()

	_, _, err = client.Collection("emails_summaries").Add(ctx, EmailSummary{
		TimeDate: getDateTime(),
		Summary:  response,
	})

	if err != nil {
		log.Printf("An error occured: %s", err)
	}

	return err
}

type NotificationBody struct {
	Message string `json:"message"`
}

func SendNotification(firebase *firebase.App) {
	ctx := context.Background()
	client, err := firebase.Messaging(ctx)
	if err != nil {
		log.Fatalf("Error getting Messaging client: %v\n", err)
	}

	topic := "new_summary"
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: "👀📈 Watchdog Alert",
			Body:  "New Email summary Alert from your Favourite Watchdog",
		},
		Topic: topic,
	}

	// Send the message
	response, err := client.Send(ctx, message)
	if err != nil {
		log.Fatalf("error sending messaging: %v\n", err)
	}

	log.Printf("Successfully sent message: %s\n", response)

}
