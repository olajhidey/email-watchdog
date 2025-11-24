# Go Email Listener & AI Summarizer

This Go application acts as a real-time email monitoring service. It connects to a Gmail account, watches a specific label for new emails from a designated sender, and uses an external AI service to summarize the email's content into simple, understandable terms.

## Features

*   **Real-time Monitoring**: Uses IMAP IDLE to get instant notifications for new emails without constant polling.
*   **Targeted Filtering**: Monitors a specific Gmail label (e.g., "INBOX", "Notifications") and filters for emails from a specific sender.
*   **AI-Powered Summarization**: Extracts the body of new emails and sends it to a compatible AI API (like OpenAI) for summarization.
*   **Cloud Persistence**: Stores AI-generated summaries in Firebase Firestore for data persistence and future retrieval.
*   **Push Notifications**: Sends alerts via Firebase Cloud Messaging to subscribed clients when new summaries are processed.
*   **Secure Configuration**: Keeps all sensitive credentials (passwords, API keys) outside the source code using a `.env` file.
*   **MIME Parsing**: Correctly parses multipart emails to extract the plain text or HTML body for processing.

## Prerequisites

*   [Go](https://go.dev/doc/install) (version 1.25 or newer)
*   A Gmail account.
*   A **Gmail App Password**. You cannot use your regular Gmail password. See Google's documentation on how to [Sign in with App Passwords](https://support.google.com/accounts/answer/185833).
*   Access to an AI service that is compatible with the OpenAI Chat Completions API format.
*   A Firebase project with Firestore enabled and a service account key (credentials.json).

## Setup & Installation

1.  **Clone the repository:**
    ```sh
    git clone <your-repository-url>
    cd email-detector
    ```

2.  **Set up Firebase credentials:**
    - Go to the [Firebase Console](https://console.firebase.google.com/), create a new project (or use an existing one).
    - Enable Firestore Database in your project.
    - Navigate to Project Settings > Service accounts > Generate new private key.
    - Download the JSON file and rename it to `credentials.json`, placing it in the root directory of this project.

3.  **Create the configuration file:**
    Copy the provided `.env.example` file to `.env` in the root of the project and fill in the required values for your environment. This file will store your secret credentials.

    ```env
    # Gmail Credentials
    GMAIL_USER="your-email@gmail.com"
    GMAIL_APP_PASSWORD="your-gmail-app-password"
    
    # Email Filtering
    TARGET_EMAIL="sender-to-watch@example.com"
    GMAIL_LABEL="INBOX"
    
    # AI Service Configuration
    AI_URL="https://api.openai.com/v1/chat/completions"
    AI_API_KEY="your-ai-api-key"
    AI_MODEL="gpt-4-turbo"
    ```

    **Configuration Details:**
    *   `GMAIL_USER`: Your full Gmail address.
    *   `GMAIL_APP_PASSWORD`: The 16-character App Password generated from your Google account.
    *   `TARGET_EMAIL`: The sender's email address you want to monitor.
    *   `GMAIL_LABEL`: The Gmail label to watch (e.g., `INBOX`, `MyLabel`, `Projects/Urgent`).
    *   `AI_URL`: The endpoint for the AI chat completion service.
    *   `AI_API_KEY`: The API key for your AI service.
    *   `AI_MODEL`: The model identifier (e.g., `gpt-4-turbo`, `gpt-3.5-turbo`).

4.  **Install dependencies:**
    Run the following command to download the necessary Go modules.
    ```sh
    go mod tidy
    ```

## How to Run

Execute the program from your terminal:

```sh
go run main.go
```

The application will connect to Gmail and start listening. When a new email from your `TARGET_EMAIL` arrives in the specified `GMAIL_LABEL`, you will see log output in your terminal, followed by the AI-generated summary.

## Running with Docker

To run the application in a Docker container:

1. **Build the Docker image:**
   ```sh
   docker build -t email-detector .
   ```

2. **Run the container:**
   ```sh
   docker run --env-file .env email-detector
   ```

This will start the application in a container, using the environment variables from your `.env` file.

Note: Ensure Docker is installed and running on your system.

## How It Works

1.  **Connect & Login**: The application establishes a secure TLS connection to Gmail's IMAP server and logs in using the provided credentials.
2.  **Select Label**: It selects the mailbox corresponding to the `GMAIL_LABEL`.
3.  **Start IDLE**: It enters the IMAP `IDLE` state, which allows the server to push updates to the client in real-time.
4.  **Wait for Update**: The program waits for a `MailboxUpdate`, which signals a new email has arrived.
5.  **Stop IDLE & Fetch**: It gracefully stops the `IDLE` command and fetches the envelope (metadata) and body of the newly arrived email.
6.  **Filter & Parse**: It checks if the email's sender matches the `TARGET_EMAIL`. If it does, it parses the email's MIME structure to get the main text content.
7.  **Summarize and Store**: The text content is sent to the configured AI API via a POST request, the response is stored in Firebase Firestore for persistence, and displayed in the console.
8.  **Restart**: The application restarts the `IDLE` command to wait for the next email.
