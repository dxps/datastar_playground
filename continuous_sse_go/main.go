package main

import (
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	datastar "github.com/starfederation/datastar/sdk/go"
)

//go:embed home.html
var homeHTML []byte

func main() {

	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(homeHTML)
	})

	r.Get("/stream", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("stream"))
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		sse := datastar.NewSSE(w, r)
		for {
			select {
			case <-r.Context().Done():
				slog.Info("Client connection closed.")
				return
			case <-ticker.C:
				bytes := make([]byte, 3)

				if _, err := rand.Read(bytes); err != nil {
					slog.Error("Failed to generate random bytes.", "error", err)
					return
				}
				hexString := hex.EncodeToString(bytes)
				frag := fmt.Sprintf(
					`<span id="feed" style="color:#%s;border:1px solid #%s;border-radius:0.25rem;padding:1rem;">%s</span>`,
					hexString, hexString, hexString)

				if err := sse.MergeFragments(frag); err != nil {
					slog.Error("Failed to merge fragments.", "error", err)
				}
			}
		}
	})

	slog.Info("Listening on :8080 ...")
	_ = http.ListenAndServe(":8080", r)
}
