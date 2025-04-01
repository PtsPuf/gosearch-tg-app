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
	Breaches []string `json:"breaches"`        // Найденные утечки
	Error    string   `json:"error,omitempty"` // Сообщение об ошибке
}

// Handler is the main entry point for Vercel serverless function
func Handler(w http.ResponseWriter, r *http.Request) {
	// Log the incoming request path and method for debugging
	log.Printf("Received request: Method=%s, Path=%s, URL=%s", r.Method, r.URL.Path, r.URL.String())

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
			log.Println("Error: TELEGRAM_BOT_TOKEN environment variable not set")
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		// First try to get chat info
		chatURL := fmt.Sprintf("https://api.telegram.org/bot%s/getChat?chat_id=@%s", token, username)
		chatResp, err := client.Get(chatURL)
		if err != nil {
			log.Printf("Error making request to Telegram API: %v", err)
			http.Error(w, fmt.Sprintf("Error making request: %v", err), http.StatusInternalServerError)
			return
		}
		defer chatResp.Body.Close()

		// Read response body
		chatBody, err := io.ReadAll(chatResp.Body)
		if err != nil {
			log.Printf("Error reading Telegram API response: %v", err)
			http.Error(w, fmt.Sprintf("Error reading response: %v", err), http.StatusInternalServerError)
			return
		}

		// Parse response
		var chatResult map[string]interface{}
		if err := json.Unmarshal(chatBody, &chatResult); err != nil {
			log.Printf("Error parsing Telegram API response: %v", err)
			http.Error(w, fmt.Sprintf("Error parsing response: %v", err), http.StatusInternalServerError)
			return
		}

		// Check if the chat exists
		if okValue, okType := chatResult["ok"].(bool); !okType || !okValue {
			// If chat not found, try to search for the user
			searchURL := fmt.Sprintf("https://api.telegram.org/bot%s/searchChatMembers?chat_id=@%s&query=%s", token, username, username)
			searchResp, err := client.Get(searchURL)
			if err != nil {
				log.Printf("Error searching for user: %v", err)
				http.Error(w, fmt.Sprintf("Error searching for user: %v", err), http.StatusInternalServerError)
				return
			}
			defer searchResp.Body.Close()

			searchBody, err := io.ReadAll(searchResp.Body)
			if err != nil {
				log.Printf("Error reading search response: %v", err)
				http.Error(w, fmt.Sprintf("Error reading search response: %v", err), http.StatusInternalServerError)
				return
			}

			var searchResult map[string]interface{}
			if err := json.Unmarshal(searchBody, &searchResult); err != nil {
				log.Printf("Error parsing search response: %v", err)
				http.Error(w, fmt.Sprintf("Error parsing search response: %v", err), http.StatusInternalServerError)
				return
			}

			// Create a proper response
			result := SearchResult{
				Username: username,
				FoundOn:  []string{},
				Breaches: []string{},
			}

			if okValue, okType := searchResult["ok"].(bool); okType && okValue {
				if members, ok := searchResult["result"].([]interface{}); ok && len(members) > 0 {
					result.FoundOn = append(result.FoundOn, "Telegram")
				} else {
					result.Error = "Пользователь не найден в Telegram"
				}
			} else {
				result.Error = "Пользователь не найден в Telegram"
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)
			return
		}

		// If chat exists, return success
		result := SearchResult{
			Username: username,
			FoundOn:  []string{"Telegram"},
			Breaches: []string{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
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
