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
	handler := http.HandlerFunc(mainHandle)
	requests := []struct {
		request string
		want    int
	}{
		{"/cafe?city=moscow&count=0", 0},
		{"/cafe?city=moscow&count=1", 1},
		{"/cafe?city=tula&count=2", 2},
		{"/cafe?city=moscow&count=100", min(100, len(cafeList["moscow"]))},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)

		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)

		body, _ := io.ReadAll(response.Body)
		bodyString := strings.TrimSpace(string(body))

		var cafes []string
		if bodyString != "" {
			cafes = strings.Split(bodyString, ",")
		} else {
			cafes = []string{}
		}

		assert.Equal(t, v.want, len(cafes))

	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/cafe?city=moscow&search=%s", v.search), nil)

		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)

		body, _ := io.ReadAll(response.Body)
		bodyString := strings.TrimSpace(string(body))

		var cafes []string
		if bodyString != "" {
			cafes = strings.Split(bodyString, ",")
		} else {
			cafes = []string{}
		}

		for _, cafeCurrent := range cafes {
			assert.True(t, strings.Contains(strings.ToLower(cafeCurrent), strings.ToLower(v.search)))
		}

		assert.Equal(t, v.wantCount, len(cafes))

	}
}
