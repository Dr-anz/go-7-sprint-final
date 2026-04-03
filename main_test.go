package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	// Создаём обработчик на основе нашей функции mainHandle
	// Это позволяет тестировать логику без запуска реального HTTP‑сервера
	handler := http.HandlerFunc(mainHandle)

	// Тестовая таблица: пары «входные данные — ожидаемый результат»
	requests := []struct {
		count int // значение параметра count, которое передаём в запросе
		want  int // ожидаемое количество кафе в ответе
	}{
		{count: 0, want: 0},                         // если count=0, ждём 0 кафе
		{count: 1, want: 1},                         // если count=1, ждём 1 кафе
		{count: 2, want: 2},                         // если count=2, ждём 2 кафе
		{count: 100, want: len(cafeList["moscow"])}, // если count большой, ждём все кафе Москвы
	}

	// Перебираем все тестовые случаи
	for _, req := range requests {
		// Формируем URL с параметрами city и count
		// Например: "/cafe?city=moscow&count=2"
		url := fmt.Sprintf("/cafe?city=moscow&count=%d", req.count)

		// Создаём «записывающее устройство» для ответа сервера
		// Оно будет хранить ответ (статус, тело и т. д.), который вернёт наш обработчик
		response := httptest.NewRecorder()

		// Создаём имитированный HTTP‑запрос
		// Метод GET, указанный URL, тело запроса пустое (nil)
		httpReq := httptest.NewRequest("GET", url, nil)

		// Запускаем обработчик: передаём ему запрос и записывающее устройство
		// Обработчик выполнит логику mainHandle и запишет результат в response
		handler.ServeHTTP(response, httpReq)

		// Проверяем, что сервер вернул статус 200 OK
		require.Equal(t, http.StatusOK, response.Code)

		// Читаем тело ответа — это строка с названиями кафе через запятую
		body := response.Body.String()

		// Разбиваем строку на слайс строк по запятым
		cafes := strings.Split(body, ",")

		// Обрабатываем случай пустой строки
		// Если в ответе ничего нет, Split вернёт [""] — заменяем на пустой слайс []string{}
		if len(cafes) == 1 && cafes[0] == "" {
			cafes = []string{}
		}

		// Сравниваем количество кафе в ответе с ожидаемым значением
		assert.Equal(t, req.want, len(cafes))
	}
}

func TestCafeSearch(t *testing.T) {
	// Создаём обработчик на основе нашей функции mainHandle
	// Это позволяет тестировать логику без запуска реального HTTP‑сервера
	handler := http.HandlerFunc(mainHandle)

	// Тестовая таблица для поиска кафе
	requests := []struct {
		search    string // поисковая подстрока
		wantCount int    // ожидаемое количество найденных кафе
	}{
		{search: "фасоль", wantCount: 0}, // «фасоль» не встречается — ждём 0 результатов
		{search: "кофе", wantCount: 2},   // «кофе» есть в 2 кафе Москвы
		{search: "вилка", wantCount: 1},  // «вилка» есть в 1 кафе Москвы
	}

	// Перебираем тестовые случаи
	for _, req := range requests {
		// Формируем URL с параметрами city и search
		// Например: "/cafe?city=moscow&search=кофе"
		url := fmt.Sprintf("/cafe?city=moscow&search=%s", req.search)

		// Имитируем HTTP‑ответ
		response := httptest.NewRecorder()

		// Имитируем HTTP‑запрос
		httpReq := httptest.NewRequest("GET", url, nil)

		// Запускаем обработчик с имитированным запросом
		handler.ServeHTTP(response, httpReq)

		// Проверяем статус ответа — должен быть 200 OK
		require.Equal(t, http.StatusOK, response.Code)

		// Получаем тело ответа
		body := response.Body.String()

		// Разбиваем на слайс по запятым
		cafes := strings.Split(body, ",")

		// Обрабатываем пустой ответ
		if len(cafes) == 1 && cafes[0] == "" {
			cafes = []string{}
		}

		// Проверяем количество найденных кафе
		assert.Equal(t, req.wantCount, len(cafes))

		// Дополнительно: для каждого найденного кафе проверяем, что оно содержит поисковую подстроку
		searchLower := strings.ToLower(req.search) // приводим поиск к нижнему регистру
		for _, cafe := range cafes {
			// Утверждаем, что название кафе содержит подстроку (без учёта регистра)
			assert.True(t, strings.Contains(strings.ToLower(cafe), searchLower))
		}
	}
}
