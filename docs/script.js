// Инициализация Telegram WebApp
const tg = window.Telegram.WebApp;
tg.ready(); // Сообщаем ТГ, что приложение готово
tg.expand(); // Раскрываем приложение на весь экран

// Получаем элементы DOM
const usernameInput = document.getElementById('usernameInput');
const searchButton = document.getElementById('searchButton');
const loader = document.getElementById('loader');
const resultsContainer = document.getElementById('results');
const statusConsole = document.getElementById('statusConsole');
const matrixBg = document.getElementById('matrixBg');

// URL вашего бэкенда
const BACKEND_URL = 'https://gosearch-tg-app.vercel.app'; // Vercel URL
// const BACKEND_URL = 'http://localhost:8080'; // Для локального теста

// Эффект "матричного дождя"
function setupMatrixBackground() {
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');
    
    // Размер матрицы на весь экран
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
    
    // Добавляем canvas в div матричного фона
    matrixBg.appendChild(canvas);
    
    // Массив для хранения символов (капель)
    let drops = [];
    
    // Символы для использования (японская катакана и цифры для кибер-эффекта)
    const matrix = "01アイウエオカキクケコサシスセソタチツテトナニヌネノハヒフヘホマミムメモヤユヨラリルレロワン";
    
    // Размер шрифта и количество колонок
    const fontSize = 14;
    const columns = Math.floor(canvas.width / fontSize);
    
    // Инициализация массива капель
    for(let i = 0; i < columns; i++) {
        drops[i] = Math.floor(Math.random() * canvas.height);
    }
    
    // Основная функция рисования матрицы
    function draw() {
        // Полупрозрачный черный фон для создания эффекта хвоста
        ctx.fillStyle = 'rgba(0, 0, 0, 0.05)';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        
        // Зеленый цвет для символов
        ctx.fillStyle = '#00ff00';
        ctx.font = fontSize + 'px monospace';
        
        // Рисуем символы
        for(let i = 0; i < drops.length; i++) {
            // Случайный символ для вывода
            const text = matrix[Math.floor(Math.random() * matrix.length)];
            
            // x-координата каждой капли, y-координата из массива
            ctx.fillText(text, i * fontSize, drops[i] * fontSize);
            
            // Если капля достигла дна, вероятностно отправляем её обратно наверх
            if(drops[i] * fontSize > canvas.height && Math.random() > 0.975) {
                drops[i] = 0;
            }
            
            // Перемещаем каплю вниз
            drops[i]++;
        }
    }
    
    // Запускаем анимацию
    setInterval(draw, 35);
    
    // Обновляем размер при изменении окна
    window.addEventListener('resize', () => {
        canvas.width = window.innerWidth;
        canvas.height = window.innerHeight;
        const newColumns = Math.floor(canvas.width / fontSize);
        
        // Обновляем массив капель если изменилось количество колонок
        if (newColumns !== columns) {
            drops.length = 0;
            for(let i = 0; i < newColumns; i++) {
                drops[i] = Math.floor(Math.random() * canvas.height);
            }
        }
    });
}

// Эффект печатающей машинки
function typeMessage(element, message, speed = 30) {
    let i = 0;
    element.innerHTML = '';
    
    function typeChar() {
        if (i < message.length) {
            element.innerHTML += message.charAt(i);
            i++;
            setTimeout(typeChar, speed);
        } else {
            // Добавляем моргающий курсор в конце сообщения
            element.innerHTML += ' <span class="cursor">_</span>';
        }
    }
    
    typeChar();
}

// Добавление сообщения в консоль
function addConsoleMessage(message) {
    const consoleLines = statusConsole.getElementsByClassName('console-line');
    
    // Используем последнюю строку если она уже заполнена или создаем новую
    let lastLine = consoleLines[consoleLines.length - 1];
    
    if (lastLine && lastLine.innerText.trim() !== '') {
        // Создаем новую строку
        const newLine = document.createElement('div');
        newLine.className = 'console-line';
        statusConsole.appendChild(newLine);
        
        // Печатаем сообщение
        typeMessage(newLine, message);
        
        // Скроллим консоль вниз
        statusConsole.scrollTop = statusConsole.scrollHeight;
    } else {
        // Используем существующую строку
        typeMessage(lastLine, message);
    }
}

// Функция для отображения результатов
function displayResults(data) {
    resultsContainer.innerHTML = ''; // Очищаем предыдущие результаты
    resultsContainer.style.display = 'block'; // Показываем контейнер

    // Добавляем сообщение в консоль
    addConsoleMessage(`АНАЛИЗ ЗАВЕРШЕН: ${data.username}`);

    // Создаем заголовок с эффектом печати
    const titleDiv = document.createElement('h2');
    resultsContainer.appendChild(titleDiv);
    typeMessage(titleDiv, `РЕЗУЛЬТАТЫ СКАНИРОВАНИЯ: ${escapeHtml(data.username)}`, 20);

    // Небольшая задержка для эффекта
    setTimeout(() => {
        let htmlContent = '';

        if (data.error && !data.found_on?.length && !data.breaches?.length) {
            // Если есть ошибка и нет других данных, показываем ее
            htmlContent += `<p class="error-message">${escapeHtml(data.error)}</p>`;
            addConsoleMessage("ОШИБКА: Объект не обнаружен в реестрах");
        } else {
            // Раздел найденных сайтов
            if (data.found_on && data.found_on.length > 0) {
                htmlContent += `<div class="result-section">
                                <h3>ОБНАРУЖЕНЫ ЦИФРОВЫЕ СЛЕДЫ (${data.found_on.length}):</h3>
                                <ul>`;
                data.found_on.forEach(site => {
                    htmlContent += `<li>${escapeHtml(site)}</li>`;
                });
                htmlContent += `</ul></div>`;
                addConsoleMessage(`НАЙДЕНО ПРОФИЛЕЙ: ${data.found_on.length}`);
            } else {
                htmlContent += `<p>ЦИФРОВЫЕ СЛЕДЫ НЕ ОБНАРУЖЕНЫ В ПУБЛИЧНОМ ДОСТУПЕ.</p>`;
                addConsoleMessage("СТАТУС: Цифровой след не обнаружен");
            }

            // Раздел утечек
            if (data.breaches && data.breaches.length > 0) {
                htmlContent += `<div class="result-section">
                                <h3>ВНИМАНИЕ! ОБНАРУЖЕНЫ УТЕЧКИ ДАННЫХ</h3>
                                <p class="warning-message">КОМПРОМЕТАЦИЯ ДАННЫХ ПОДТВЕРЖДЕНА. РЕКОМЕНДУЕТСЯ НЕМЕДЛЕННАЯ СМЕНА УЧЕТНЫХ ДАННЫХ НА ВСЕХ СВЯЗАННЫХ РЕСУРСАХ!</p>
                                <ul>`;
                data.breaches.forEach(breachInfo => {
                    htmlContent += `<li>${escapeHtml(breachInfo)}</li>`;
                });
                htmlContent += `</ul></div>`;
                addConsoleMessage(`УГРОЗА: Обнаружено ${data.breaches.length} утечек`);
            } else {
                htmlContent += `<p>УТЕЧЕК ДАННЫХ НЕ ОБНАРУЖЕНО В БАЗАХ CYBERCRIME.</p>`;
                addConsoleMessage("БЕЗОПАСНОСТЬ: Утечек не выявлено");
            }

            // Общее заключение
            if (!data.breaches || data.breaches.length === 0) {
                if (data.found_on && data.found_on.length > 0) {
                    htmlContent += `<p class="success-message">УТЕЧЕК НЕ НАЙДЕНО, ЦИФРОВОЙ СЛЕД ОГРАНИЧЕН</p>`;
                } else {
                    htmlContent += `<p class="success-message">ПРОФИЛЬ ЧИСТ. ЦИФРОВЫЕ СЛЕДЫ И УТЕЧКИ НЕ ОБНАРУЖЕНЫ.</p>`;
                }
            }
        }

        // Добавляем HTML в контейнер
        const contentDiv = document.createElement('div');
        contentDiv.innerHTML = htmlContent;
        resultsContainer.appendChild(contentDiv);
    }, 800); // Задержка для лучшего визуального эффекта
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
        addConsoleMessage("ОШИБКА: Не указана цель сканирования");
        tg.showAlert('Укажите имя пользователя для анализа.');
        return;
    }

    // Показываем прелоадер и скрываем старые результаты
    loader.style.display = 'block';
    resultsContainer.style.display = 'none';
    resultsContainer.innerHTML = '';
    searchButton.disabled = true; // Блокируем кнопку на время запроса

    // Анимируем процесс поиска в консоли
    addConsoleMessage(`ИНИЦИАЛИЗАЦИЯ ПОИСКА: ${username}`);
    setTimeout(() => addConsoleMessage("ПОДКЛЮЧЕНИЕ К БАЗАМ ДАННЫХ..."), 500);
    setTimeout(() => addConsoleMessage("ЗАПРОС ОТПРАВЛЕН К API CYBERCRIME..."), 1200);
    setTimeout(() => addConsoleMessage("СКАНИРОВАНИЕ СОЦИАЛЬНЫХ СЕТЕЙ..."), 2000);

    try {
        const response = await fetch(`${BACKEND_URL}/search?username=${encodeURIComponent(username)}`);

        if (!response.ok) {
            // Попытка прочитать тело ошибки, если бэкенд его отдает
            let errorText = `СЕТЕВАЯ ОШИБКА: ${response.status} ${response.statusText}`;
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
        setTimeout(() => {
            addConsoleMessage("АНАЛИЗ ДАННЫХ ЗАВЕРШЕН");
            displayResults(data);
        }, 2500);

    } catch (error) {
        console.error('Ошибка при выполнении запроса:', error);
        setTimeout(() => {
            addConsoleMessage(`КРИТИЧЕСКАЯ ОШИБКА: ${error.message.substring(0, 50)}...`);
            resultsContainer.style.display = 'block';
            resultsContainer.innerHTML = `<p class="error-message">СБОЙ АНАЛИЗА: ${escapeHtml(error.message)}</p>`;
            tg.showAlert(`Сбой сканирования: ${error.message}`);
        }, 2500);
    } finally {
        // Скрываем прелоадер и разблокируем кнопку с небольшой задержкой для эффекта
        setTimeout(() => {
            loader.style.display = 'none';
            searchButton.disabled = false;
        }, 2500);
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

// Инициализируем матричный фон
setupMatrixBackground();

// Начальное сообщение в консоли
addConsoleMessage("СИСТЕМА АКТИВИРОВАНА");
setTimeout(() => addConsoleMessage("ГОТОВ К ПОИСКУ ЦИФРОВЫХ СЛЕДОВ"), 1000);

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