package main

import (
	"fmt"
	"net/http"
)

func main() {
	RunServer()
}

// RunServer - функция для настройки и запуска HTTP-сервера.
func RunServer() {
	// Обработчик запросов, который отправляет "OK" в ответ.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
