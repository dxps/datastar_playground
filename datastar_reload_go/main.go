package main

import (
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	datastar "github.com/starfederation/datastar/sdk/go"
)

const (
	serverAddress = "localhost:9001"
)

//go:embed home.html
var homeHTML []byte

var hotReloadOnlyOnce sync.Once

func HotReloadHandler(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)
	hotReloadOnlyOnce.Do(func() {
		// Refresh the client page as soon as connection is established.
		// This will occur only once after the server starts.
		if err := sse.ExecuteScript(
			"window.location.reload()",
			datastar.WithExecuteScriptRetryDuration(time.Second),
		); err != nil {
			slog.Error("Failed to execute script.", "error", err)
		}
	})
	// Freeze the event stream until the connection
	// is lost for any reason. This will force the client
	// to attempt to reconnect after the server reboots.
	<-r.Context().Done()
}

func PageWithHotReload(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write(homeHTML)
}

func main() {
	http.HandleFunc("/hotreload", HotReloadHandler)
	http.HandleFunc("/", PageWithHotReload)
	slog.Info(fmt.Sprintf(
		"Open your browser to: http://%s/",
		serverAddress,
	))
	_ = http.ListenAndServe(serverAddress, nil)
}
