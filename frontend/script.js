// ========================================================
// ПЕРЕКЛЮЧАТЕЛЬ: тестовый режим (без бэкенда)
// false — запросы к реальному API (localhost:8080)
// true  — имитация ответа через setTimeout
// ========================================================
const TEST_MODE = false;


// Ждём, пока вся HTML-структура загрузится, прежде чем искать элементы
document.addEventListener('DOMContentLoaded', function () {
    // --- Получаем все нужные элементы со страницы ---
    const container = document.getElementById('fileInputsContainer'); // Контейнер, куда добавляются поля файлов
    const addBtn = document.getElementById('addFileBtn');             // Кнопка "+ Добавить файл"
    const startBtn = document.getElementById('startAnalysisBtn');    // Кнопка "Let's Go"
    const loadingState = document.getElementById('loadingState');    // Блок с лоадером ("Анализ...")
    const resultsContainer = document.getElementById('resultsContainer'); // Контейнер для результатов
    const resetBtn = document.getElementById('resetBtn');           // Кнопка "Начать заново"
    const uploadFrame = document.getElementById('uploadFrame');     // Первый фрейм (с загрузкой файлов)

    // --- Переменные состояния ---
    let fileCount = 0;              // Счётчик для генерации уникальных ID
    let currentFilesData = [];      // Массив объектов { id, name, file } — храним выбранные файлы


    // ========================================================
    // 1. СОЗДАНИЕ ПОЛЯ ДЛЯ ВЫБОРА ФАЙЛА
    // ========================================================
    function createFileInput() {
        fileCount++;
        const uniqueId = 'file-' + fileCount; // Уникальный ID для этого поля

        // --- Создаём обёртку для одного поля ---
        const wrapper = document.createElement('div');
        wrapper.className = 'file-input-wrapper';

        // --- Видимое текстовое поле (readOnly — пользователь не может печатать сам) ---
        const textInput = document.createElement('input');
        textInput.type = 'text';
        textInput.className = 'file-input-field';
        textInput.placeholder = 'Выберите файл (.txt)';
        textInput.readOnly = true;

        // --- Скрытый реальный input для выбора файла ---
        const realFileInput = document.createElement('input');
        realFileInput.type = 'file';
        realFileInput.accept = '.txt';          // Ограничиваем только текстовыми файлами
        realFileInput.style.display = 'none';   // Скрываем с экрана

        // --- Кнопка удаления поля ("−") ---
        const removeBtn = document.createElement('button');
        removeBtn.type = 'button';
        removeBtn.className = 'remove-btn';
        removeBtn.innerHTML = '&minus;';
        removeBtn.title = 'Удалить файл';

        // --- Логика: клик по текстовому полю → открываем системный выбор файла ---
        textInput.addEventListener('click', function () {
            realFileInput.click();
        });

        // --- Логика: пользователь выбрал файл в проводнике ---
        realFileInput.addEventListener('change', function (e) {
			fields = document.getElementsByClassName("file-input-field")
            fields[0].style.border = "3px solid rgba(95, 158, 160, 0.461)";

            var filesList = e.target.files;        // Получаем список выбранных файлов
            if (filesList.length > 0) {
                var file = filesList[0];           // Берём первый (и единственный) файл
                textInput.value = file.name;       // Подставляем имя файла в видимое поле

                // Проверяем, есть ли уже запись для этого поля в массиве
                var index = currentFilesData.findIndex(function (item) {
                    return item.id === uniqueId;
                });

                // Если запись есть — перезаписываем, если нет — добавляем новую
                if (index !== -1) {
                    currentFilesData[index] = { id: uniqueId, name: file.name, file: file };
                } else {
                    currentFilesData.push({ id: uniqueId, name: file.name, file: file });
                }
            }
        });

        // --- Логика: удаление поля по кнопке "−" ---
        removeBtn.addEventListener('click', function () {
            // Удаляем файл из массива по ID
            currentFilesData = currentFilesData.filter(function (f) {
                return f.id !== uniqueId;
            });
            // Удаляем сам блок из DOM
            wrapper.remove();

            // Если удалили последнее поле — создаём одно пустое обратно
            if (container.children.length === 0) {
                createFileInput();
            }
        });

        // --- Собираем обёртку и добавляем на страницу ---
        wrapper.appendChild(textInput);
        wrapper.appendChild(realFileInput);
        wrapper.appendChild(removeBtn);
        container.appendChild(wrapper);
    }


    // ========================================================
    // 2. ИНИЦИАЛИЗАЦИЯ
    // ========================================================
    addBtn.addEventListener('click', createFileInput); // Кнопка "+" → новое поле
    createFileInput();                                  // Создаём первое поле сразу при загрузке


    // ========================================================
    // 3. КНОПКА "LET'S GO" — ЗАПУСК АНАЛИЗА
    // ========================================================
    startBtn.addEventListener('click', async function () {
        // --- Проверяем, выбран ли хоть один файл ---
        var inputs = container.querySelectorAll('.file-input-field');
        var hasFiles = false;
        inputs.forEach(function (input) {
            if (input.value !== '') hasFiles = true;
        });

        if (!hasFiles) {
			fields = document.getElementsByClassName("file-input-field")
            fields[0].style.border = "2px solid red";
            return; // Прерываем функцию, не запускаем анализ
        }

        // --- Блокируем первый фрейм и кнопку старта ---
        uploadFrame.classList.add('blocked');   // CSS: opacity 0.5 + pointer-events none
        startBtn.style.display = 'none';
        startBtn.style.opacity = '0.5';

        // --- Показываем лоадер, скрываем старые результаты ---
        loadingState.style.display = 'block';
        resultsContainer.style.display = 'none';
        resetBtn.style.display = 'none';


        // ====================================================
        // РЕЖИМ 1: ТЕСТОВЫЙ (без бэкенда)
        // ====================================================
        if (TEST_MODE) {
            setTimeout(function () {
                // Формируем мок-ответ по той же структуре, что и реальный сервер
                var resultsArray = currentFilesData.map(function (fileData) {
                    return {
                        'file-path': fileData.name,
                        'lines': Math.floor(Math.random() * 200) + 10,
                        'symbols': Math.floor(Math.random() * 5000) + 100,
                        'sentences': Math.floor(Math.random() * 50) + 2,
                        'words': Math.floor(Math.random() * 500) + 20,
                        'avg-word-len': (Math.random() * 5 + 3).toFixed(2),
                        'longest-word': {
                            'Text': 'экстраординарный',
                            'Length': 16,
                            'Quantity': 1
                        },
                        'top-frequents': [
                            { 'Text': 'и', 'Length': 1, 'Quantity': 45 },
                            { 'Text': 'в', 'Length': 1, 'Quantity': 32 },
                            { 'Text': 'не', 'Length': 2, 'Quantity': 28 }
                        ],
                        'err': null
                    };
                });

                renderResults(resultsArray);
            }, 1500);
            return;
        }

        // ====================================================
        // РЕЖИМ 2: РЕАЛЬНЫЙ ЗАПРОС К API
        // ====================================================
        var formData = new FormData();
        currentFilesData.forEach(function (fileData) {
            formData.append('file', fileData.file, fileData.name);
        });

         try {
            var response = await fetch('/api/analyze/multiple', {
                method: 'POST',
                body: formData
            });

            // --- Если сервер вернул HTTP-ошибку ---
            if (!response.ok) {
                // Текст ошибки читаем только для логов, на фронт не показываем
                var errorText = await response.text();
                console.error('Сервер вернул ошибку:', response.status, errorText);

                showErrorPage(response.status);
                return;
            }

            var results = await response.json();
            var resultsArray = Array.isArray(results) ? results : [results];
            renderResults(resultsArray);

        } catch (error) {
            // Сетевая ошибка (сервер недоступен, CORS, обрыв связи)
            console.error('Ошибка запроса:', error);

            // Если не получилось достать статус — считаем, что сервер недоступен (503)
            var code = 503;
            showErrorPage(code);
        }
    });


    // ========================================================
    // 3.5 ВЫВОД ОШИБКИ ВМЕСТО РЕЗУЛЬТАТОВ
    // ========================================================

    // Словарь кодов: код → название
    var HTTP_CODES = {
        400: 'Bad Request',
        401: 'Unauthorized',
        403: 'Forbidden',
        404: 'Not Found',
        405: 'Method Not Allowed',
        408: 'Request Timeout',
        413: 'Payload Too Large',
        415: 'Unsupported Media Type',
        422: 'Unprocessable Entity',
        429: 'Too Many Requests',
        500: 'Internal Server Error',
        502: 'Bad Gateway',
        503: 'Service Unavailable',
        504: 'Gateway Timeout'
    };

    function showErrorPage(code) {
        loadingState.style.display = 'none';
        resultsContainer.innerHTML = '';
        resultsContainer.style.display = 'flex';
        resultsContainer.style.backgroundColor = "#242424";
        resetBtn.style.display = 'block';

        // Название кода из словаря, либо "Unknown Error"
        var title = HTTP_CODES[code] || 'Unknown Error';

        // Создаём блок с большим числом
        var errorBlock = document.createElement('div');
        errorBlock.className = 'error-block';

        errorBlock.innerHTML =
            '<div class="error-code">' + code + '</div>' +
            '<div class="error-title">' + title + '</div>';

        resultsContainer.appendChild(errorBlock);
    }


    // ========================================================
    // 3. РЕНДЕР РЕЗУЛЬТАТОВ (общий для обоих режимов)
    // ========================================================
    function renderResults(resultsArray) {
        loadingState.style.display = 'none';
        resultsContainer.innerHTML = '';
        resultsContainer.style.display = 'flex';

        resultsArray.forEach(function (res) {
            var resultDiv = document.createElement('div');
            resultDiv.className = 'result-item';

            // --- Если сервер вернул ошибку для конкретного файла ---
            if (res['err'] && res['err'] !== null) {
                resultDiv.innerHTML =
                    '<div class="result-info">' +
                    '<strong>Файл:</strong> ' + (res['file-path'] || 'Неизвестно') + '<br>' +
                    '<strong style="color: #ff6b6b;">Ошибка:</strong> ' + res['err'] +
                    '</div>';
                resultDiv.style.borderLeftColor = "red"
                resultsContainer.appendChild(resultDiv);
                return;
            }

            // --- Текстовая версия для .txt файла ---
            // Формат точно как в консольной версии на Go
            var textContent =
                'Файл: ' + res['file-path'] + '\n' +
                'Строк: ' + res['lines'] + '\n' +
                'Символов: ' + res['symbols'] + '\n' +
                'Предложений: ' + res['sentences'] + '\n' +
                'Слов: ' + res['words'] + '\n' +
                'Средняя длина слова: ' + res['avg-word-len'] + '\n' +
                'Самое длинное слово: ' + (res['longest-word'] ? res['longest-word'].Text : '—') +
                ' - ' + (res['longest-word'] ? res['longest-word'].Length : '—') + '\n' +
                'Топ частых слов:\n' + formatTopFrequentsText(res['top-frequents']);

            // --- HTML-версия для страницы ---
            var infoDiv = document.createElement('div');
            infoDiv.className = 'result-info';
            infoDiv.innerHTML =
                '<strong>Файл:</strong> ' + res['file-path'] + '<br>' +
                '<strong>Строк:</strong> ' + res['lines'] + '<br>' +
                '<strong>Символов:</strong> ' + res['symbols'] + '<br>' +
                '<strong>Предложений:</strong> ' + res['sentences'] + '<br>' +
                '<strong>Слов:</strong> ' + res['words'] + '<br>' +
                '<strong>Средняя длина слова:</strong> ' + res['avg-word-len'] + '<br>' +
                '<strong>Самое длинное слово:</strong> ' +
                (res['longest-word'] ? res['longest-word'].Text + ' — ' + res['longest-word'].Length : '—') + '<br>' +
                '<strong>Топ частых слов:</strong><br>' +
                formatTopFrequentsHTML(res['top-frequents']);

            // --- Кнопка скачивания ---
            var downloadBtn = document.createElement('button');
            downloadBtn.type = 'button';
            downloadBtn.className = 'download-btn';
            downloadBtn.innerHTML = '📥 Скачать анализ';

            // Имя файла: берём file-path, убираем путь, оставляем только имя
            var shortName = res['file-path'].split('/').pop().split('\\').pop();
            downloadBtn.addEventListener('click', function () {
                downloadTextFile(textContent, shortName + '_analysis.txt');
            });

            resultDiv.appendChild(infoDiv);
            resultDiv.appendChild(downloadBtn);
            resultsContainer.appendChild(resultDiv);
        });

        resetBtn.style.display = 'block';
    }


    // ========================================================
    // 4. ФОРМАТИРОВАНИЕ ЧАСТЫХ СЛОВ
    // ========================================================

    // Для HTML — каждый на отдельной строке с отступом
    function formatTopFrequentsHTML(top) {
        if (!top || top.length === 0) return '&nbsp;&nbsp;Нет данных';
        var lines = top.map(function (w) {
            return '&nbsp;|&nbsp;' + w.Text + ': ' + w.Quantity;
        });
        return lines.join('<br>');
    }

    // Для текстового файла — каждый с новой строки с отступом
    function formatTopFrequentsText(top) {
        if (!top || top.length === 0) return '  Нет данных\n';
        var lines = top.map(function (w) {
            return ' | ' + w.Text + ': ' + w.Quantity;
        });
        return lines.join('\n') + '\n';
    }


    // ========================================================
    // 4. КНОПКА "НАЧАТЬ ЗАНОВО" — СБРОС ИНТЕРФЕЙСА
    // ========================================================
    resetBtn.addEventListener('click', function () {
        // Очищаем результаты
        resultsContainer.innerHTML = '';
        resultsContainer.style.display = 'none';

        // Скрываем лоадер (на всякий случай)
        loadingState.style.display = 'none';

        // Разблокируем первый фрейм и кнопку старта
        uploadFrame.classList.remove('blocked');
        startBtn.style.display = 'block';
        startBtn.style.opacity = '1';

        // Скрываем саму кнопку сброса
        resetBtn.style.display = 'none';

        // Очищаем массив файлов
        currentFilesData = [];

        // Сбрасываем поля ввода: оставляем одно, удаляем остальные
        var wrappers = container.querySelectorAll('.file-input-wrapper');
        wrappers.forEach(function (w, i) {
            if (i > 0) {
                w.remove(); // Лишние поля — удаляем
            } else {
                // Первое поле — очищаем значения
                var textInput = w.querySelector('.file-input-field');
                var realInput = w.querySelector('input[type="file"]');
                if (textInput) textInput.value = '';
                if (realInput) realInput.value = ''; // Без этого браузер не даст выбрать тот же файл повторно
            }
        });
    });


    // ========================================================
    // 5. ВСПОМОГАТЕЛЬНАЯ ФУНКЦИЯ: СКАЧИВАНИЕ ТЕКСТОВОГО ФАЙЛА
    // ========================================================
    function downloadTextFile(text, filename) {
        // Создаём "виртуальный" файл из текста
        var blob = new Blob([text], { type: 'text/plain;charset=utf-8' });
        var url = URL.createObjectURL(blob);

        // Создаём временную ссылку и кликаем по ней
        var a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();

        // Удаляем ссылку и освобождаем память
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }
});