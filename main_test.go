package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gopkg.in/resty.v1"
)

// TestRun выполняет тестирование функции RunWithContext.
func TestRun(t *testing.T) {
	// Создание тестового сервера с обработчиком, который возвращает "OK"
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	}))
	defer testServer.Close()
	// Создание объекта конфигурации с использованием URL тестового сервера
	config := &Config{
		URL: testServer.URL,
		Requests: struct {
			Amount    int `mapstructure:"amount"`
			PerSecond int `mapstructure:"per_second"`
		}{
			Amount:    10,
			PerSecond: 2,
		},
	}
	// Создание логгера
	logger, _ := NewLogger()
	// Создание клиента Resty
	client := resty.New()
	// Создание контекста и функции отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Создание канала для сигнала о завершении выполнения
	done := make(chan struct{})
	// Запуск функции RunWithContext в отдельной горутине
	go RunWithContext(ctx, config, logger, client, done)
	// Закрытие канала done, чтобы уведомить, что выполнение завершено
	close(done)
	<-time.After(5 * time.Second)
	// Ожидание сигнала о завершении выполнения функции
	select {
	case <-done:
		// Если сигнал получен, тест пройден успешно
	case <-time.After(3 * time.Second):
		// Если сигнал не получен в течение 3 секунд, тест завершается с ошибкой
		t.Error("The test has timed out")
	}
}
