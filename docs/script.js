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
        // Поставим summary сверху для быстрого обзора
        let summaryContent = '';
        
        // Счетчик проверенных источников - используем реальное значение из ответа
        const totalSitesChecked = data.total_sites_checked || 0;
        const sitesCountText = totalSitesChecked > 0 ? `${totalSitesChecked}` : "300+";
        summaryContent += `<p><span class="blink-badge success"></span>ПРОВЕРЕНО ${sitesCountText} ИСТОЧНИКОВ</p>`;
        
        // Проверка на ошибки
        if (data.error && !data.found_on?.length && !data.breaches?.length) {
            summaryContent += `<p class="error-message">${escapeHtml(data.error)}</p>`;
            addConsoleMessage("ОШИБКА: Объект не обнаружен в реестрах");
        } else {
            // Результаты и общее заключение
            if (data.found_on && data.found_on.length > 0) {
                summaryContent += `<p><span class="blink-badge warning"></span>ОБНАРУЖЕНО ${data.found_on.length} ЦИФРОВЫХ СЛЕДОВ</p>`;
            } else {
                summaryContent += `<p><span class="blink-badge success"></span>ЦИФРОВЫЕ СЛЕДЫ НЕ ОБНАРУЖЕНЫ</p>`;
            }
            
            if (data.breaches && data.breaches.length > 0) {
                summaryContent += `<p><span class="blink-badge danger"></span>ОБНАРУЖЕНО ${data.breaches.length} УТЕЧЕК ПАРОЛЕЙ</p>`;
            } else {
                summaryContent += `<p><span class="blink-badge success"></span>УТЕЧКИ ПАРОЛЕЙ НЕ ОБНАРУЖЕНЫ</p>`;
            }
            
            // Общее заключение
            if (!data.breaches || data.breaches.length === 0) {
                if (data.found_on && data.found_on.length > 0) {
                    summaryContent += `<p class="success-message">УТЕЧЕК НЕ НАЙДЕНО, ЦИФРОВОЙ СЛЕД ОГРАНИЧЕН</p>`;
                } else {
                    summaryContent += `<p class="success-message">ПРОФИЛЬ ЧИСТ. ЦИФРОВЫЕ СЛЕДЫ И УТЕЧКИ НЕ ОБНАРУЖЕНЫ.</p>`;
                }
            }
        }
        
        // Добавляем summary
        const summaryDiv = document.createElement('div');
        summaryDiv.className = 'summary-section';
        summaryDiv.innerHTML = summaryContent;
        resultsContainer.appendChild(summaryDiv);
        
        // Если нет ошибок и есть результаты, добавляем секции с деталями
        if (!data.error || data.found_on?.length || data.breaches?.length) {
            
            // Секция найденных сайтов
            if (data.found_on && data.found_on.length > 0) {
                const sitesSection = createCollapsibleSection(
                    `ЦИФРОВЫЕ СЛЕДЫ (${data.found_on.length})`, 
                    'warning', 
                    true // Открыто по умолчанию
                );
                
                // Создаем сетку для сайтов
                const sitesGrid = document.createElement('div');
                sitesGrid.className = 'sites-grid';
                
                // Добавляем каждый сайт в сетку
                data.found_on.forEach(site => {
                    const siteItem = document.createElement('div');
                    siteItem.className = 'site-item';
                    siteItem.textContent = site;
                    sitesGrid.appendChild(siteItem);
                });
                
                // Добавляем сетку в содержимое секции
                sitesSection.querySelector('.collapsible-content').appendChild(sitesGrid);
                resultsContainer.appendChild(sitesSection);
                
                addConsoleMessage(`НАЙДЕНО ПРОФИЛЕЙ: ${data.found_on.length}`);
            }
            
            // Секция утечек
            if (data.breaches && data.breaches.length > 0) {
                const breachesSection = createCollapsibleSection(
                    `УТЕЧКИ ПАРОЛЕЙ (${data.breaches.length})`, 
                    'danger',
                    true // Открыто по умолчанию
                );
                
                // Предупреждение
                const warningMsg = document.createElement('p');
                warningMsg.className = 'warning-message';
                warningMsg.textContent = 'КОМПРОМЕТАЦИЯ ПАРОЛЕЙ ПОДТВЕРЖДЕНА. РЕКОМЕНДУЕТСЯ НЕМЕДЛЕННАЯ СМЕНА УЧЕТНЫХ ДАННЫХ!';
                breachesSection.querySelector('.collapsible-content').appendChild(warningMsg);
                
                // Список утечек
                const breachesList = document.createElement('ul');
                data.breaches.forEach(breachInfo => {
                    const item = document.createElement('li');
                    item.textContent = breachInfo;
                    breachesList.appendChild(item);
                });
                
                breachesSection.querySelector('.collapsible-content').appendChild(breachesList);
                resultsContainer.appendChild(breachesSection);
                
                addConsoleMessage(`УГРОЗА: Обнаружено ${data.breaches.length} утечек паролей`);
            } else {
                addConsoleMessage("БЕЗОПАСНОСТЬ: Утечек паролей не выявлено");
            }
        }
        
    }, 800); // Задержка для лучшего визуального эффекта
}

// Функция для создания сворачиваемой секции
function createCollapsibleSection(title, badgeType = null, isOpenByDefault = false) {
    const section = document.createElement('div');
    section.className = 'collapsible-section';
    if (isOpenByDefault) section.className += ' open';
    
    const header = document.createElement('div');
    header.className = 'collapsible-header';
    if (isOpenByDefault) header.className += ' active';
    
    // Добавляем индикатор если указан тип
    if (badgeType) {
        const badge = document.createElement('span');
        badge.className = `blink-badge ${badgeType}`;
        header.appendChild(badge);
    }
    
    const headerText = document.createElement('span');
    headerText.textContent = title;
    header.appendChild(headerText);
    
    const toggle = document.createElement('span');
    toggle.className = 'collapsible-toggle';
    toggle.textContent = '›';
    header.appendChild(toggle);
    
    const content = document.createElement('div');
    content.className = 'collapsible-content';
    
    section.appendChild(header);
    section.appendChild(content);
    
    // Добавляем обработчик клика
    header.addEventListener('click', () => {
        section.classList.toggle('open');
        header.classList.toggle('active');
    });
    
    return section;
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

    // Анимируем процесс поиска в консоли с увеличенным временем ожидания
    addConsoleMessage(`ИНИЦИАЛИЗАЦИЯ ПОИСКА: ${username}`);
    setTimeout(() => addConsoleMessage("ПОДКЛЮЧЕНИЕ К БАЗАМ ДАННЫХ..."), 800);
    setTimeout(() => addConsoleMessage("ЗАПРОС ОТПРАВЛЕН К API CYBERCRIME..."), 2000);
    setTimeout(() => addConsoleMessage("СКАНИРОВАНИЕ СОЦИАЛЬНЫХ СЕТЕЙ..."), 3500);
    setTimeout(() => addConsoleMessage("ПРОВЕРКА НАЛИЧИЯ ПРОФИЛЕЙ..."), 5000);
    setTimeout(() => addConsoleMessage("АНАЛИЗ БАЗ ДАННЫХ УТЕЧЕК ПАРОЛЕЙ..."), 7000);
    setTimeout(() => addConsoleMessage("ВЫПОЛНЕНИЕ ДЕТАЛЬНОЙ ПРОВЕРКИ..."), 9000);
    setTimeout(() => addConsoleMessage("ФОРМИРОВАНИЕ ЦИФРОВОГО СЛЕДА..."), 12000);

    try {
        // Создаем контроллер для управления таймаутом
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), 40000); // 40 секунд таймаут
        
        const response = await fetch(
            `${BACKEND_URL}/search?username=${encodeURIComponent(username)}`,
            { signal: controller.signal }
        );
        
        clearTimeout(timeoutId); // Очищаем таймаут если запрос завершился успешно

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