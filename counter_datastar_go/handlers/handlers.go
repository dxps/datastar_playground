package handlers

import (
	"net/http"

	"github.com/dxps/datastar_playground/counter_datastar_go/internal/components"
)

func HomeGetHandler(w http.ResponseWriter, r *http.Request) {
	RenderView(w, r, components.HomeView("hello, world!"), "/")
}

func CounterGetHandler(w http.ResponseWriter, r *http.Request) {

	// sessionManager, ok := r.Context().Value(middleware.SessionManagerKey).(*scs.SessionManager)
	// if !ok {
	// 	onError(w, errors.New("Couldnt get context value (sessionManager)"),
	// 		"Internal server error", http.StatusInternalServerError)
	// }

	// userCount := sessionManager.GetInt(r.Context(), "count")
	// g := models.GlobalValuesInstance{ID: 1}
	// err := g.Read(dbClient)
	// onError(w, err, "Internal server error", http.StatusInternalServerError)

	RenderView(w, r, components.CounterView(g.Count, userCount), "/counter")
}
