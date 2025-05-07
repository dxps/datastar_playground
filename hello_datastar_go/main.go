package main

import (
	_ "embed"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	datastar "github.com/starfederation/datastar/sdk/go"
)

//go:embed hello-world.html
var helloWorldHTML []byte

func main() {
	r := chi.NewRouter()

	const message = "Hello, world!"
	type Store struct {
		Delay time.Duration `json:"delay"` // delay in milliseconds between each character of the message.
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(helloWorldHTML)
	})

	r.Get("/hello-world", func(w http.ResponseWriter, r *http.Request) {
		store := &Store{}
		if err := datastar.ReadSignals(r, store); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		sse := datastar.NewSSE(w, r)

		for i := 0; i < len(message); i++ {
			if err := sse.MergeFragments(`<div id="message">` + message[:i+1] + `</div>`); err != nil {
				slog.Error("Failed on MergeFragments", "error", err)
			}
			time.Sleep(store.Delay * time.Millisecond)
		}
	})

	slog.Info("Listening on :9001 ...")
	if err := http.ListenAndServe(":9001", r); err != nil {
		slog.Error("Failed to listen", "error", err)
	}
}
