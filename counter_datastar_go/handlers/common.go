package handlers

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/dxps/datastar_playground/counter_datastar_go/internal/components"
)

func RenderView(w http.ResponseWriter, r *http.Request, view templ.Component, layoutPath string) {

	// TBD
	if r.Header.Get("Hx-Request") == "true" {
		err := view.Render(r.Context(), w)
		onError(w, err, "Internal server error", http.StatusInternalServerError)
		return
	}

	err := components.Layout(layoutPath).Render(r.Context(), w)
	onError(w, err, "Internal server error", http.StatusInternalServerError)

}

func onError(w http.ResponseWriter, err error, msg string, code int) {
	if err != nil {
		http.Error(w, msg, code)
		slog.Error(msg, "error", err)
	}
}
