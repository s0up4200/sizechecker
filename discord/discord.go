package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func SendDiscordNotification(webhookURL, message string) error {
	payloadBytes, err := json.Marshal(map[string]string{"content": message})
	if err != nil {
		return fmt.Errorf("error marshalling JSON payload: %v", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("received non-204 response from Discord: %s - %s", resp.Status, string(bodyBytes))
	}

	return nil
}
