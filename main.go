package main

import (
	"io"
	"log"
	"os"
	"strings"

	firebase "firebase.google.com/go"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/joho/godotenv"
	"github.com/olajhidey/gofetch/utils"
)

func main() {
	// load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v\n", err)
	}

	firebaseApp, err := utils.LoadFirebaseConfig()
	if err != nil {
		log.Fatalf("Error loading Firebase config: %v", err)
	}
	log.Println("Firebase initialized successfully")

	username := os.Getenv("GMAIL_USER")
	password := os.Getenv("GMAIL_APP_PASSWORD")
	targetEmail := os.Getenv("TARGET_EMAIL")
	label := os.Getenv("GMAIL_LABEL")

	// Connect to Gmail
	c, err := client.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		log.Fatalf("Error connecting to Gmail: %v", err)
	}
	defer c.Logout()

	// Login
	if err := c.Login(username, password); err != nil {
		log.Fatalf("Error logging in: %v", err)
	}

	log.Println("Connected to Gmail successfully!")

	// Further email processing logic would go here...
	_, err = c.Select(label, false) // <-- Change "INBOX" to your desired label name
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Label %s selected. Waiting for emails from: %s\n", label, targetEmail)

	updates := make(chan client.Update)
	c.Updates = updates

	// This channel is used to signal the IDLE command to stop
	stop := make(chan struct{})

	// This channel is used to get the error result from the IDLE command
	done := make(chan error, 1)

	go func() {
		done <- c.Idle(stop, nil)
	}()

	// Start listening for new emails
	for {
		log.Println("Listening for new emails...")
		select {
		case update := <-updates:
			if mboxUpdate, ok := update.(*client.MailboxUpdate); ok {
				log.Println("New email detected! Checking sender...")
				newEmailSeqNum := mboxUpdate.Mailbox.Messages
				log.Println("New email is message number:", newEmailSeqNum)

				// Stop the IDLE command gracefully
				close(stop)
				<-done // Wait for the IDLE command to finish

				fetchNewEmails(c, newEmailSeqNum, targetEmail, firebaseApp)

				// Restart IDLE with a new stop channel
				stop = make(chan struct{})
				go func() { done <- c.Idle(stop, nil) }()
			}

		case err := <-done:
			// IDLE terminated unexpectedly
			if err != nil {
				log.Fatal(err)
			}
			log.Println("IDLE finished")
			return
		}
	}
}

func fetchNewEmails(c *client.Client, seqNum uint32, targetEmail string, firebase *firebase.App) {
	if seqNum == 0 {
		log.Println("Invalid sequence number for new email.")
		return
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(seqNum)

	log.Println(seqset.String())

	// Define what parts of the email we want to fetch
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, section.FetchItem()}
	messages := make(chan *imap.Message, 10)
	fetchDone := make(chan error, 1)
	go func() {
		fetchDone <- c.Fetch(seqset, items, messages)
	}()

	for {
		select {
		case msg := <-messages:
			if msg == nil {
				log.Println("Message stream ended.")
				return
			}
			sender := msg.Envelope.From[0].Address()
			subject := msg.Envelope.Subject

			if strings.Contains(strings.ToLower(sender), strings.ToLower(targetEmail)) {
				log.Printf("New email from target %s: %s\n", targetEmail, subject)
				r := msg.GetBody(section)
				if r == nil {
					log.Println("Server didn't return message body")
					continue
				}
				bodyContent := parseMIME(r)
				responseFromAi := utils.ResponseTextFromAi(bodyContent)
				log.Println("AI summary successful")

				// Implementation to send FCM notification
				// utils.SendNotification(firebase)

				// Store response from AI to cloud store or something
				err := utils.StoreResponseToDatabase(firebase, responseFromAi)
				if err != nil {
					log.Println("Something went wrong: ", err)
				}
				log.Println("Database insertion successful")
			}
		case err := <-fetchDone:
			if err != nil && err != io.EOF {
				log.Printf("Fetch error: %v", err)
			}
			return
		}
	}
}

func parseMIME(r io.Reader) string {

	var textBody, htmlBody, finalBody string

	// Create a  mail reader to parse MIME multiparts messages
	mr, err := mail.CreateReader(r)
	if err != nil {
		log.Printf("Failed to parse email MIME: %v:", err)
		return "Error parsing email"
	}

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading email part: %v", err)
			break
		}

		contentType := p.Header.Get("Content-Type")
		separate := strings.Split(contentType, ";")
		contentType = strings.TrimSpace(separate[0])
		log.Printf("Processing part with Content-Type: %s", contentType)

		switch contentType {
		case "text/plain":
			bodyBytes, _ := io.ReadAll(p.Body)
			textBody = string(bodyBytes)
		case "text/html":
			bodyBytes, _ := io.ReadAll(p.Body)
			htmlBody = string(bodyBytes)
		}
	}

	if textBody != "" {
		finalBody = textBody
	} else if htmlBody != "" {
		finalBody = htmlBody
	}

	return finalBody
}
