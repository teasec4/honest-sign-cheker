package main

import (
	"flag"
	"fmt"
	"net/http"
	"onestsignt/internal/handler"
	"os"

)

const maxUploadSize = 512 << 20

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "адрес API-сервера")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handler.HandleHealth)
	mux.HandleFunc("/api/primary-check", handler.HandlePrimaryCheck)
	mux.HandleFunc("/api/duplicate-check", handler.HandleDuplicateCheck)

	fmt.Printf("API-сервер запущен: http://%s\n", *addr)
	if err := http.ListenAndServe(*addr, handler.WithCORS(mux)); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
		os.Exit(1)
	}
}

