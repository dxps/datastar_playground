package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
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
	CountKey   = "count"
	SessionKey = "session_count"
)

var (
	global         GlobalState
	sessionManager *scs.SessionManager
)

func getHandler(w http.ResponseWriter, r *http.Request) {
	sessCount := sessionManager.GetInt32(r.Context(), CountKey)
	signals := comps.CountersSignals{
		Global:  global.Count,
		Session: sessCount,
	}
	component := comps.Page(signals)
	_ = component.Render(r.Context(), w)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()         // Update state.
	if r.Form.Has("global") { // If Global button was pressed.
		global.Count++
	}
	if r.Form.Has("session") { // If Session button was pressed.
		currentCount := sessionManager.GetInt(r.Context(), "count")
		sessionManager.Put(r.Context(), "count", currentCount+1)
	}
	getHandler(w, r) // Display the form.
}

func main() {
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	mux := http.NewServeMux()

	// Handle POST and GET requests.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			postHandler(w, r)
			return
		}
		getHandler(w, r)
	})

	//
	mux.HandleFunc("/counter/increment/global", func(w http.ResponseWriter, r *http.Request) {
		upd := gabs.New()
		global.Count++
		if _, err := upd.Set(global, "global"); err != nil {
			slog.Error("Failed to update global", "error", err)
		}
		if err := datastar.NewSSE(w, r).MarshalAndMergeSignals(upd); err != nil {
			slog.Error("Failed to MarshalAndMergeSignals", "error", err)
		}
		slog.Info("Updated global count", "value", global.Count)
	})

	// Include the static content.
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	// Add the middleware.
	muxWithSessionMiddleware := sessionManager.LoadAndSave(mux)

	// Start the server.
	fmt.Println("Listening on :9000 ...")
	if err := http.ListenAndServe("127.0.0.1:9000", muxWithSessionMiddleware); err != nil {
		log.Printf("error listening: %v", err)
	}
}
