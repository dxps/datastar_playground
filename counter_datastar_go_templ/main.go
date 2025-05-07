package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Jeffail/gabs/v2"
	"github.com/alexedwards/scs/v2"
	"github.com/dxps/htmx_playground/counter_htmx_go_templ/comps"
	datastar "github.com/starfederation/datastar/sdk/go"
)

type GlobalState struct {
	Count int32
}

const (
	CountKey        = "count"
	SessionCountKey = "session_count"
)

var (
	global         GlobalState
	sessionManager *scs.SessionManager
)

func main() {

	slog.SetDefault(
		slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		),
	)

	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		sc := sessionManager.GetInt32(r.Context(), SessionCountKey)
		signals := comps.CountersSignals{
			Global:  global.Count,
			Session: sc,
		}
		sessionManager.Put(r.Context(), SessionCountKey, sc)
		slog.Debug(fmt.Sprintf("On initial access, session count is %d.", sc))
		component := comps.Page(signals)
		_ = component.Render(r.Context(), w)
	})

	mux.HandleFunc("/counter/increment/global", func(w http.ResponseWriter, r *http.Request) {

		global.Count++
		slog.Debug(fmt.Sprintf("Updated global count to %d.", global.Count))
		upd := gabs.New()
		if _, err := upd.Set(global.Count, "global"); err != nil {
			slog.Error("Failed to update global count", "error", err)
		}
		if err := datastar.NewSSE(w, r).MarshalAndMergeSignals(upd); err != nil {
			slog.Error("Failed to MarshalAndMergeSignals w/ global count update", "error", err)
		}
	})

	mux.HandleFunc("/counter/increment/session", func(w http.ResponseWriter, r *http.Request) {

		if !sessionManager.Exists(r.Context(), SessionCountKey) {
			slog.Warn(fmt.Sprintf("%s key not found in the current session", SessionCountKey))
		}
		sc := sessionManager.GetInt32(r.Context(), SessionCountKey)
		usc := sc + 1
		sessionManager.Put(r.Context(), SessionCountKey, usc)
		slog.Debug(fmt.Sprintf("Updated session count to %d.", usc))
		upd := gabs.New()
		if _, err := upd.Set(usc, "session"); err != nil {
			slog.Error("Failed to update session", "error", err)
		}
		if err := datastar.NewSSE(w, r).MarshalAndMergeSignals(upd); err != nil {
			slog.Error("Failed to MarshalAndMergeSignals w/ session count update", "error", err)
		}
	})

	// Include the static assets.
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	// Add the middleware.
	muxWithSessionMiddleware := sessionManager.LoadAndSave(mux)

	// Start the server.
	fmt.Println("Listening on :9000 ...")
	if err := http.ListenAndServe("127.0.0.1:9000", muxWithSessionMiddleware); err != nil {
		slog.Error("Failed to listen.", "error", err)
	}
}
