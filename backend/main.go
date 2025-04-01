package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Структура для ответа API
type SearchResult struct {
	Username string   `json:"username"`
	FoundOn  []string `json:"found_on"`        // Сайты, где найден пользователь
	Breaches []string `json:"breaches"`        // Найденные утечки (пока просто как пример)
	Error    string   `json:"error,omitempty"` // Сообщение об ошибке, если есть
}

// Handler is the main entry point for Vercel serverless function
func Handler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Handle search endpoint
	if r.URL.Path == "/search" {
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "Username parameter is required", http.StatusBadRequest)
			return
		}

		// Get Telegram API token from environment (handled by Vercel)
		token := os.Getenv("TELEGRAM_BOT_TOKEN")
		if token == "" {
			// Log the error server-side for debugging
			log.Println("Error: TELEGRAM_BOT_TOKEN environment variable not set")
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		// Make request to Telegram API
		url := fmt.Sprintf("https://api.telegram.org/bot%s/getChat?chat_id=@%s", token, username)
		resp, err := client.Get(url)
		if err != nil {
			log.Printf("Error making request to Telegram API: %v", err) // Log error
			http.Error(w, fmt.Sprintf("Error making request: %v", err), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading Telegram API response: %v", err) // Log error
			http.Error(w, fmt.Sprintf("Error reading response: %v", err), http.StatusInternalServerError)
			return
		}

		// Parse response
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			log.Printf("Error parsing Telegram API response: %v", err) // Log error
			http.Error(w, fmt.Sprintf("Error parsing response: %v", err), http.StatusInternalServerError)
			return
		}

		// Check if the chat exists - fixed variable shadowing
		if okValue, okType := result["ok"].(bool); !okType || !okValue {
			// Check if there's an error message from Telegram
			if description, hasDesc := result["description"].(string); hasDesc {
				log.Printf("Telegram API error for user %s: %s", username, description)
				if description == "Bad Request: chat not found" {
					http.Error(w, "Chat not found", http.StatusNotFound)
				} else {
					http.Error(w, fmt.Sprintf("Telegram API error: %s", description), http.StatusInternalServerError)
				}
			} else {
				log.Printf("Telegram API returned 'ok: false' or unexpected structure for user %s. Response: %s", username, string(body))
				http.Error(w, "Chat not found or invalid response", http.StatusNotFound)
			}
			return
		}

		// Return the successful response from Telegram
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
		return
	}

	// Handle root endpoint
	if r.URL.Path == "/" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("GoSearch Telegram API is running via Vercel"))
		return
	}

	// Handle 404 for any other paths
	http.Error(w, "Not Found", http.StatusNotFound)
}
