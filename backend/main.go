package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
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

		// Get Telegram API token from environment
		token := os.Getenv("TELEGRAM_BOT_TOKEN")
		if token == "" {
			http.Error(w, "Telegram bot token not configured", http.StatusInternalServerError)
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
			http.Error(w, fmt.Sprintf("Error making request: %v", err), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading response: %v", err), http.StatusInternalServerError)
			return
		}

		// Parse response
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			http.Error(w, fmt.Sprintf("Error parsing response: %v", err), http.StatusInternalServerError)
			return
		}

		// Check if the chat exists
		if ok, ok := result["ok"].(bool); !ok || !ok {
			http.Error(w, "Chat not found", http.StatusNotFound)
			return
		}

		// Return the response
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
		return
	}

	// Handle root endpoint
	if r.URL.Path == "/" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("GoSearch Telegram API is running"))
		return
	}

	// Handle 404
	http.Error(w, "Not Found", http.StatusNotFound)
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, http.HandlerFunc(Handler)); err != nil {
		log.Fatal(err)
	}
}
