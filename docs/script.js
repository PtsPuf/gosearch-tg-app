// Инициализация Telegram WebApp
const tg = window.Telegram.WebApp;
tg.ready(); // Сообщаем ТГ, что приложение готово
tg.expand(); // Раскрываем приложение на весь экран

// Получаем элементы DOM
const usernameInput = document.getElementById('usernameInput');
const searchButton = document.getElementById('searchButton');
const loader = document.getElementById('loader');
const resultsContainer = document.getElementById('results');

// URL вашего бэкенда
const BACKEND_URL = 'https://gosearch-tg-app.vercel.app'; // Updated to Vercel URL
// const BACKEND_URL = 'http://localhost:8080'; // Для локального теста

// Функция для отображения результатов
function displayResults(data) {
    resultsContainer.innerHTML = ''; // Очищаем предыдущие результаты
    resultsContainer.style.display = 'block'; // Показываем контейнер

    let htmlContent = `<h2>Результаты для: ${escapeHtml(data.username)}</h2>`;

    if (data.error && !data.found_on?.length && !data.breaches?.length) {
        // Если есть ошибка и нет других данных, показываем ее
        htmlContent += `<p class="error-message">${escapeHtml(data.error)}</p>`;
    } else {
        // Раздел найденных сайтов
        if (data.found_on && data.found_on.length > 0) {
            htmlContent += `<div class="result-section">
                              <h3>Найден на сайтах (${data.found_on.length}):</h3>
                              <ul>`;
            data.found_on.forEach(site => {
                htmlContent += `<li>${escapeHtml(site)}</li>`;
            });
            htmlContent += `</ul></div>`;
        } else {
            htmlContent += `<p>Профили на отслеживаемых сайтах не найдены.</p>`;
        }

        // Раздел утечек
        if (data.breaches && data.breaches.length > 0) {
            htmlContent += `<div class="result-section">
                              <h3>Обнаружены возможные утечки!</h3>
                              <p class="warning-message">Обнаружены признаки того, что данные, связанные с этим именем пользователя, могли быть скомпрометированы. <strong>Рекомендуется сменить пароли</strong> на всех связанных аккаунтах!</p>
                              <ul>`;
            data.breaches.forEach(breachInfo => {
                // Пока просто выводим сообщение из бэкенда
                htmlContent += `<li>${escapeHtml(breachInfo)}</li>`;
            });
            htmlContent += `</ul></div>`;
        } else {
            htmlContent += `<p>Признаков утечки данных для этого пользователя не найдено.</p>`;
        }

        // Общее заключение
        if (!data.breaches || data.breaches.length === 0) {
            if (data.found_on && data.found_on.length > 0) {
                 htmlContent += `<p class="success-message">Утечек не найдено, но профили обнаружены. Все в порядке!</p>`;
            } else {
                htmlContent += `<p class="success-message">Профили и утечки не найдены. Все чисто!</p>`;
            }
        }
    }

    resultsContainer.innerHTML = htmlContent;
}

// Функция для экранирования HTML
function escapeHtml(unsafe) {
    if (!unsafe) return '';
    return unsafe
         .replace(/&/g, "&amp;")
         .replace(/</g, "&lt;")
         .replace(/>/g, "&gt;")
         .replace(/"/g, "&quot;")
         .replace(/'/g, "&#039;");
 }

// Функция для выполнения поиска
async function performSearch() {
    const username = usernameInput.value.trim();
    if (!username) {
        // Можно добавить уведомление через tg.showAlert()
        tg.showAlert('Пожалуйста, введите имя пользователя.');
        return;
    }

    // Показываем прелоадер и скрываем старые результаты
    loader.style.display = 'block';
    resultsContainer.style.display = 'none';
    resultsContainer.innerHTML = '';
    searchButton.disabled = true; // Блокируем кнопку на время запроса

    try {
        const response = await fetch(`${BACKEND_URL}/search?username=${encodeURIComponent(username)}`);

        if (!response.ok) {
            // Попытка прочитать тело ошибки, если бэкенд его отдает
            let errorText = `Ошибка сети: ${response.status} ${response.statusText}`;
            try {
                 const errorData = await response.json();
                 errorText = errorData.error || JSON.stringify(errorData);
            } catch (e) {
                // Ошибка парсинга JSON, используем текстовый ответ
                try {
                    errorText = await response.text();
                } catch (e2) {
                    // Не удалось прочитать текст ошибки
                }
            }
            throw new Error(errorText);
        }

        const data = await response.json();
        displayResults(data);

    } catch (error) {
        console.error('Ошибка при выполнении запроса:', error);
        resultsContainer.style.display = 'block';
        resultsContainer.innerHTML = `<p class="error-message">Не удалось выполнить поиск: ${escapeHtml(error.message)}</p>`;
        tg.showAlert(`Ошибка поиска: ${error.message}`);
    } finally {
        // Скрываем прелоадер и разблокируем кнопку
        loader.style.display = 'none';
        searchButton.disabled = false;
    }
}

// Обработчик клика по кнопке
searchButton.addEventListener('click', performSearch);

// Обработчик нажатия Enter в поле ввода
usernameInput.addEventListener('keypress', function(event) {
    if (event.key === 'Enter') {
        performSearch();
    }
});

// --- Дополнительные возможности Telegram WebApp (по желанию) ---

// Можно изменить цвет шапки
// tg.setHeaderColor('#ff0000');

// Показать кнопку "Назад"
// tg.BackButton.show();
// tg.BackButton.onClick(() => {
//     window.history.back(); // Или другое действие
// });

// Показать основную кнопку
// tg.MainButton.setText('Закрыть');
// tg.MainButton.show();
// tg.MainButton.onClick(() => {
//     tg.close();
// }); 