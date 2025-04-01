package handler

import (
	"context"
	_ "embed" // Импортируем для //go:embed
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Структура для сайта из data.json
type SiteInfo struct {
	Name      string      `json:"name"`
	BaseURL   string      `json:"base_url"`
	URLProbe  string      `json:"url_probe"` // Используем, если есть, для проверки
	ErrorType string      `json:"errorType"`
	ErrorCode interface{} `json:"errorCode"` // Может быть int или string
	ErrorMsg  string      `json:"errorMsg"`
	// Добавим поле для User-Agent, если понадобится
	// UserAgent string `json:"user_agent,omitempty"`
}

// Встраиваем data.json в переменную sitesData
// Путь "data.json" здесь относителен к файлу main.go
//
//go:embed data.json
var sitesData []byte

// Глобальная переменная для хранения данных сайтов
var sites []SiteInfo
var once sync.Once // Для однократной загрузки data.json

// Функция для загрузки data.json
func loadSites() {
	once.Do(func() {
		// Читаем из встроенной переменной sitesData
		if err := json.Unmarshal(sitesData, &sites); err != nil {
			log.Fatalf("Error unmarshalling embedded data.json: %v", err)
		}
		if len(sites) == 0 {
			log.Fatalf("Embedded data.json seems empty or invalid")
		}
		log.Printf("Loaded %d sites from embedded data.json", len(sites))
	})
}

// Структура для ответа API
type SearchResult struct {
	Username string   `json:"username"`
	FoundOn  []string `json:"found_on"`        // Сайты, где найден пользователь
	Breaches []string `json:"breaches"`        // Найденные утечки (пока не используется)
	Error    string   `json:"error,omitempty"` // Сообщение об ошибке
}

// Функция проверки одного сайта
func checkSite(ctx context.Context, client *http.Client, site SiteInfo, username string, resultsChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	checkURL := site.BaseURL
	if site.URLProbe != "" {
		checkURL = site.URLProbe // Используем URL для проверки, если он указан
	}
	targetURL := strings.Replace(checkURL, "{}", username, 1)

	// Добавляем User-Agent, чтобы имитировать браузер
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		// Не логируем ошибку создания запроса, т.к. их может быть много
		// log.Printf("Error creating request for %s: %v", site.Name, err)
		return
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		// Не логируем ошибки сети, т.к. их может быть много
		// log.Printf("Error checking %s (%s): %v", site.Name, targetURL, err)
		return
	}
	defer resp.Body.Close()

	// --- Логика проверки ---
	found := false
	switch site.ErrorType {
	case "status_code":
		// Ожидаем, что errorCode - это число (статус код ошибки)
		var expectedErrorCode int
		switch v := site.ErrorCode.(type) {
		case float64: // JSON числа часто парсятся как float64
			expectedErrorCode = int(v)
		case int:
			expectedErrorCode = v
		default:
			// Не можем обработать - пропускаем
			return
		}
		// Пользователь найден, если статус НЕ равен коду ошибки
		if resp.StatusCode != expectedErrorCode {
			found = true
		}
	case "errorMsg":
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return // Не можем прочитать тело - пропускаем
		}
		bodyString := string(bodyBytes)
		// Пользователь найден, если тело НЕ содержит сообщение об ошибке
		if !strings.Contains(bodyString, site.ErrorMsg) {
			found = true
		}
	case "profilePresence":
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return // Не можем прочитать тело - пропускаем
		}
		bodyString := string(bodyBytes)
		// Пользователь найден, если тело СОДЕРЖИТ сообщение о наличии профиля
		if strings.Contains(bodyString, site.ErrorMsg) {
			found = true
		}
	case "unknown":
		// Не можем определить - пропускаем сайт
		return
	default:
		// Неизвестный тип ошибки - пропускаем
		return
	}

	if found {
		select {
		case resultsChan <- site.Name: // Отправляем имя сайта, если нашли
		case <-ctx.Done(): // Прекращаем, если контекст завершен (например, таймаут)
			return
		}
	}
}

// Handler is the main entry point for Vercel serverless function
func Handler(w http.ResponseWriter, r *http.Request) {
	// Загружаем данные сайтов при первом вызове
	loadSites()

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
		// Увеличим общий таймаут, т.к. проверяем много сайтов
		client := &http.Client{
			Timeout: 20 * time.Second, // Общий таймаут для всех запросов к сайтам
		}

		// --- Проверка Telegram ---
		var wgTelegram sync.WaitGroup
		telegramFoundChan := make(chan bool, 1)
		wgTelegram.Add(1)
		go func() {
			defer wgTelegram.Done()
			chatURL := fmt.Sprintf("https://api.telegram.org/bot%s/getChat?chat_id=@%s", token, username)
			reqTg, _ := http.NewRequest("GET", chatURL, nil)  // Создаем запрос для добавления User-Agent
			reqTg.Header.Set("User-Agent", "GoSearchBot/1.0") // Добавляем User-Agent
			chatResp, err := client.Do(reqTg)
			if err != nil {
				log.Printf("Error making request to Telegram API getChat: %v", err)
				telegramFoundChan <- false
				return
			}
			defer chatResp.Body.Close()

			chatBody, err := io.ReadAll(chatResp.Body)
			if err != nil {
				log.Printf("Error reading Telegram API getChat response: %v", err)
				telegramFoundChan <- false
				return
			}

			var chatResult map[string]interface{}
			if err := json.Unmarshal(chatBody, &chatResult); err != nil {
				log.Printf("Error parsing Telegram API getChat response: %v", err)
				telegramFoundChan <- false
				return
			}

			if okValue, okType := chatResult["ok"].(bool); okType && okValue {
				telegramFoundChan <- true
				return
			}
			telegramFoundChan <- false
		}()

		// --- Проверка сайтов из data.json ---
		var wgSites sync.WaitGroup
		resultsChan := make(chan string, len(sites)) // Канал для имен найденных сайтов
		// Контекст с таймаутом для всех проверок сайтов
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // Уменьшил таймаут для сайтов
		defer cancel()                                                           // Важно отменить контекст

		for _, site := range sites {
			wgSites.Add(1)
			go checkSite(ctx, client, site, username, resultsChan, &wgSites)
		}

		// Горутина для ожидания завершения всех проверок сайтов
		go func() {
			wgSites.Wait()
			close(resultsChan) // Закрываем канал, когда все горутины завершились
		}()

		// Сбор результатов
		foundSites := []string{}

		// Ждем результат от Telegram
		wgTelegram.Wait()
		if <-telegramFoundChan {
			foundSites = append(foundSites, "Telegram")
		}

		// Собираем результаты от проверки сайтов
		for siteName := range resultsChan {
			foundSites = append(foundSites, siteName)
		}

		// Формируем финальный ответ
		finalResult := SearchResult{
			Username: username,
			FoundOn:  foundSites,
			Breaches: []string{}, // Пока не ищем утечки
		}

		if len(foundSites) == 0 {
			finalResult.Error = "Пользователь не найден ни на одном из проверяемых сайтов."
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(finalResult)
		return
	}

	// Handle root endpoint
	if r.URL.Path == "/" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("GoSearch Telegram API is running via Vercel with multi-site check (embedded data)"))
		return
	}

	// Handle 404 for any other paths
	http.Error(w, "Not Found", http.StatusNotFound)
}
