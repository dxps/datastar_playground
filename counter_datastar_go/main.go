package main

import (
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/Jeffail/gabs/v2"
	"github.com/dxps/datastar_playground/counter_datastar_go/handlers"
	"github.com/dxps/datastar_playground/counter_datastar_go/internal/components"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"
	datastar "github.com/starfederation/datastar/sdk/go"
)

func main() {

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	store := sessions.NewFilesystemStore("/tmp", []byte("secret"))

	if err := setupExamplesTemplCounter(r, store); err != nil {
		slog.Error("Setup failed", "err", err)
	}

	slog.Info("Listening on port 3000 ...")
	if err := http.ListenAndServe(":3000", r); err != nil {
		slog.Error("ListenAndServe failed", "err", err)
	}

}

func setupExamplesTemplCounter(router chi.Router, sessionSignals sessions.Store) error {

	fs := http.FileServer(http.Dir("static"))
	router.Handle("/static/*", http.StripPrefix("/static/", fs))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.RenderView(w, r, components.HomeView("hello, Datastar Counter!"), "/")
	})

	router.Get("/counter", func(w http.ResponseWriter, r *http.Request) {
		handlers.RenderView(w, r, components.CounterView(0, 0), "/counter")
	})

	var globalCounter atomic.Uint32
	const (
		sessionKey = "templ_counter"
		countKey   = "count"
	)

	sessFunc := func(r *http.Request) (uint32, *sessions.Session, error) {
		sess, err := sessionSignals.Get(r, sessionKey)
		if err != nil {
			return 0, nil, err
		}

		val, ok := sess.Values[countKey].(uint32)
		if !ok {
			val = 0
		}
		return val, sess, nil
	}

	router.Get("/counter/get", func(w http.ResponseWriter, r *http.Request) {
		sessVal, _, err := sessFunc(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		signals := components.TemplCounterSignals{
			Global:  globalCounter.Load(),
			Session: sessVal,
		}

		c := components.TemplCounterExampleInitialContents(signals)
		if err := datastar.NewSSE(w, r).MergeFragmentTempl(c); err != nil {
			slog.Error("Failed to merge fragment templ", "err", err)
		}
	})

	updateGlobal := func(signals *gabs.Container) {
		if _, err := signals.Set(globalCounter.Add(1), "global"); err != nil {
			slog.Error("Failed to update global counter", "err", err)
		}
	}

	router.Route("/counter/increment", func(incrementRouter chi.Router) {
		incrementRouter.Post("/global", func(w http.ResponseWriter, r *http.Request) {
			update := gabs.New()
			updateGlobal(update)

			if err := datastar.NewSSE(w, r).MarshalAndMergeSignals(update); err != nil {
				slog.Error("Failed to marshal and merge signals", "err", err)
			}
		})

		incrementRouter.Post("/session", func(w http.ResponseWriter, r *http.Request) {
			val, sess, err := sessFunc(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			val++
			sess.Values[countKey] = val
			if err := sess.Save(r, w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			update := gabs.New()
			updateGlobal(update)
			if _, err := update.Set(val, "user"); err != nil {
				slog.Error("Failed to update user counter", "err", err)
			}

			if err := datastar.NewSSE(w, r).MarshalAndMergeSignals(update); err != nil {
				slog.Error("Failed to marshal and merge signals", "err", err)
			}
		})
	})

	return nil
}
