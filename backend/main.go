package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

// Структура для ответа API
type SearchResult struct {
	Username string   `json:"username"`
	FoundOn  []string `json:"found_on"`        // Сайты, где найден пользователь
	Breaches []string `json:"breaches"`        // Найденные утечки (пока просто как пример)
	Error    string   `json:"error,omitempty"` // Сообщение об ошибке, если есть
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Параметр 'username' обязателен", http.StatusBadRequest)
		return
	}

	log.Printf("Начинаю поиск для пользователя: %s", username)

	// --- Выполнение gosearch ---
	// ПРИМЕЧАНИЕ: Путь к gosearch может потребоваться настроить
	// или убедиться, что он в PATH окружения Render.com
	// Используем --no-false-positives для большей точности
	cmd := exec.Command("gosearch", "-u", username, "--no-false-positives")
	output, err := cmd.CombinedOutput() // Получаем и stdout, и stderr

	result := SearchResult{
		Username: username,
		FoundOn:  []string{},
		Breaches: []string{}, // Инициализируем пустым срезом
	}

	if err != nil {
		// gosearch может завершиться с ошибкой, даже если что-то найдено
		// или если просто ничего не найдено. Анализируем вывод.
		log.Printf("Ошибка выполнения gosearch для %s: %v. Вывод: %s", username, err, string(output))
		// Пока просто логируем, попробуем распарсить вывод ниже
		// result.Error = fmt.Sprintf("Ошибка выполнения gosearch: %v", err)
	}

	// --- Парсинг вывода gosearch ---
	// Это примерный парсинг, его нужно будет уточнить под реальный вывод gosearch
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[+] Found user:") {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				// Убираем URL и оставляем только имя сайта
				siteInfo := strings.SplitN(parts[1], " at ", 2)
				if len(siteInfo) > 0 {
					// Очищаем от возможных ANSI кодов цвета (простой вариант)
					cleanSite := strings.Split(siteInfo[0], "\x1b")[0]
					result.FoundOn = append(result.FoundOn, strings.TrimSpace(cleanSite))
				}
			}
		} else if strings.Contains(line, "potential breach") || strings.Contains(line, "Password hash found") {
			// Пример обнаружения утечек - нужно адаптировать под реальный вывод
			result.Breaches = append(result.Breaches, "Обнаружена потенциальная утечка данных или хеш пароля!")
		}
		// TODO: Добавить парсинг других видов информации из gosearch
	}

	if len(result.FoundOn) == 0 && len(result.Breaches) == 0 && result.Error == "" {
		result.Error = "Пользователь не найден ни на одном из сайтов и утечек не обнаружено."
	}

	log.Printf("Результат поиска для %s: %+v", username, result)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Разрешаем запросы с любого источника (для GitHub Pages)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func main() {
	// Инициализируем go.mod (если еще не сделано)
	// Вам нужно будет выполнить `go mod init <имя_модуля>` и `go mod tidy` в папке backend
	// Например: `go mod init myapp/backend`

	// Проверяем, доступен ли gosearch
	_, err := exec.LookPath("gosearch")
	if err != nil {
		log.Fatal("Команда 'gosearch' не найдена в PATH. Установите gosearch: go install github.com/ibnaleem/gosearch@latest")
	} else {
		log.Println("Команда 'gosearch' найдена.")
	}

	http.HandleFunc("/search", searchHandler)

	port := "8080" // Порт по умолчанию для Render.com
	log.Printf("Сервер запускается на порту %s", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
