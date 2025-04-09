package main

import (
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestCafeWhenNegative(t *testing.T) {
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

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	for k, v := range cafeList {
		requests := []struct {
			count int
			want  int
		}{
			{0, 0},
			{1, 1},
			{2, 2},
			{100, int(math.Min(float64(len(v)), 100))},
		}
		for _, l := range requests {
			response := httptest.NewRecorder()

			params := url.Values{}
			params.Add("city", k)
			params.Add("count", strconv.Itoa(l.count))

			req := httptest.NewRequest("GET", "/?"+params.Encode(), nil)

			handler.ServeHTTP(response, req)

			body := strings.TrimSpace(response.Body.String())
			var cafes []string

			if body != "" {
				cafes = strings.Split(body, ",")
			} else {
				cafes = []string{}
			}

			assert.Equal(t, l.want, len(cafes), "for city %s and count %d, expected %d cafes, got %d",
				k, l.count, l.want, len(cafes))
			require.Equal(t, http.StatusOK, response.Code)
		}
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

	city := "moscow"
	for _, l := range requests {
		response := httptest.NewRecorder()

		params := url.Values{}
		params.Add("city", city)
		params.Add("search", l.search)

		req := httptest.NewRequest("GET", "/?"+params.Encode(), nil)

		handler.ServeHTTP(response, req)

		body := strings.TrimSpace(response.Body.String())

		var cafes []string
		if body != "" {
			cafes = strings.Split(body, ",")
		} else {
			cafes = []string{}
		}

		count := 0
		for _, v := range cafes {
			if strings.Contains(strings.ToLower(v), strings.ToLower(l.search)) {
				count++
			}
		}

		assert.Equal(t, l.wantCount, count, "for city %s and search %q, expected %d cafes, got %d",
			city, l.search, l.wantCount, count)
		require.Equal(t, http.StatusOK, response.Code)
	}
}
