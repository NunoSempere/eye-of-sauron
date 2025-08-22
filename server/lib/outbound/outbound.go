package outbound

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

func SendPostmarkEmail(html_body string) (string, error) {
	// Create the email payload using an anonymous struct
	email := "nuno.semperelh@protonmail.com"
	log.Printf("\n\nEmail body: %s\n\n", html_body)
	payload := struct {
		From          string `json:"From"`
		To            string `json:"To"`
		Subject       string `json:"Subject"`
		TextBody      string `json:"TextBody"`
		HtmlBody      string `json:"HtmlBody"`
		TrackLinks    string `json:"TrackLinks"`
		MessageStream string `json:"MessageStream"`
	}{
		From:          "server@nunosempere.com",
		To:            email,
		Subject:       "Warning",
		TextBody:      "",
		HtmlBody:      html_body,
		TrackLinks:    "None",
		MessageStream: "outbound",
	}

	// Marshal the payload into JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error unmarshalling the json payload: %v\n", err)
		return "", err
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://api.postmarkapp.com/email", bytes.NewReader(jsonData))
	if err != nil {
		log.Printf("Error making request: %v\n", err)
		return "", err
	}

	// Read key from environment
	// hopefully read from a .env gile in the parent with godotenv
	postmark_token := os.Getenv("POSTMARK_KEY")

	// Set the request headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", postmark_token)

	// Perform the HTTP request
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send email: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response: %v\n", err)
		return "", err
	}

	log.Printf("Response status: %s\n", resp.Status)
	log.Printf("Response body: %s\n", body)
	return (resp.Status + "\n" + string(body)), nil
}
