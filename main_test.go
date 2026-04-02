package main

import (
	"fmt"
	"io"
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

	// определяем текстовую таблицу
	requests := []struct {
		count int // передаваемое значение count
		want  int // ожидаемое количество кафе в ответе
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: len(cafeList["moscow"])}, // предпологаем что проверяем на москве
	}

	for _, req := range requests {
		// формируем URL с параметром count
		url := "http://localhost:8080/cafe?city=moscow&count=" + fmt.Sprintf("%d", req.count)

		// отправляем GET-запрос
		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()

		// проверяем что запрос успешно обработан (статус 200 ОК)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// читаем тело ответа
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		// разбиваем строку на слайс по запятым
		cafes := strings.Split(string(body), ",")

		// если строка пустая то слайс будет содержать один пустой элемент
		if len(cafes) == 1 && cafes[0] == "" {
			cafes = []string{}
		}

		// сравниваем длину слайса с ожидаемым количеством
		assert.Equal(t, req.want, len(cafes))
	}
}

func TestCafeSearch(t *testing.T) {

	// определяем тестовую таблицу
	requests := []struct {
		search    string // передаваемое значение search
		wantCount int    // ожидаемое количество кафе в ответе
	}{
		{search: "фасоль", wantCount: 0},
		{search: "кофе", wantCount: 2},
		{search: "вилка", wantCount: 1},
	}

	for _, req := range requests {
		// формируем URL с параметрами city и search
		url := "http://localhost:8080/cafe?city=moscow&search=" + req.search

		// отправляем GET-запрос
		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()

		// проверяем что запрос успешно обработан (статус 200 ОК)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// читаем тело ответа
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		// разбиваем строку на слайс по запятым
		cafes := strings.Split(string(body), ",")

		// обрабатываем случай пустой строки
		if len(cafes) == 1 && cafes[0] == "" {
			cafes = []string{}
		}

		// проверяем количество найденных кафе
		assert.Equal(t, req.wantCount, len(cafes))

		//для каждого кафе проверяем что оно содержит подстроку search (без учета регистра)
		searchLower := strings.ToLower(req.search)
		for _, cafe := range cafes {
			assert.True(t, strings.Contains(strings.ToLower(cafe), searchLower))
		}
	}
}
