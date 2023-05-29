package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mar-cial/items/db"
)

type errResponse struct {
	status  int    `json:"status"`
	message string `json:"message"`
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})

}

func CreateRouter(app *app) *mux.Router {
	r := mux.NewRouter()

	r.Use(commonMiddleware)

	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/items", app.listItems).Methods(http.MethodGet)
	api.HandleFunc("/sampleItems", app.createItems).Methods(http.MethodPost)

	return r
}

type app struct {
	db *db.DB
}

func NewApp() *app {
	d := db.NewDBInstance()

	return &app{
		db: d,
	}
}
func (app *app) listItems(w http.ResponseWriter, r *http.Request) {
	items, err := app.db.GetItems()

	if err != nil {
		err := &errResponse{
			status:  http.StatusInternalServerError,
			message: "err listing items",
		}

		w.WriteHeader(err.status)

		json.NewEncoder(w).Encode(err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(items)
}

func (app *app) createItems(w http.ResponseWriter, r *http.Request) {
	createRes, err := app.db.CreateItems()

	if err != nil {
		err := &errResponse{
			status:  http.StatusInternalServerError,
			message: "err listing items",
		}

		w.WriteHeader(err.status)

		json.NewEncoder(w).Encode(err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createRes)
}
