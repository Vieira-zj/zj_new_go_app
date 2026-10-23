package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

func main() {
	server := http.Server{
		Addr:              ":8081",
		Handler:           getServerMux(),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	slog.Info("http server listening at 8081")
	if err := server.ListenAndServe(); err != nil {
		panic(fmt.Errorf("server failed to listen and serve: %v", err))
	}
}

func getServerMux() *http.ServeMux {
	mux := http.NewServeMux()

	// curl http://localhost:8081/health
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintln(w, "ok"); err != nil {
			slog.Error("http write failed", slog.Any("error", err))
		}
	})
	mux.HandleFunc("/upload", uploadHandler)
	return mux
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 10 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize) // 10 MB limit read
	b, err := io.ReadAll(r.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			slog.Error(maxErr.Error(), slog.Int64("limit_size", maxErr.Limit))
			http.Error(w, "payload too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	mockedSaveFile(b)
	w.WriteHeader(http.StatusOK)
}

func mockedSaveFile(b []byte) {
	slog.Info("saved file", slog.Int("bytes", len(b)))
}
