package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/mar-cial/items/db"
	"github.com/mar-cial/items/model"
	"go.mongodb.org/mongo-driver/mongo"
)

type app struct {
	mc *mongo.Client
}

func CreateApp() (*app, error) {
	client, err := db.CreateClient()

	return &app{
		mc: client,
	}, err
}

type errResponse struct {
	message string `json:"message"`
	status  int    `json:"status"`
}

func serveErrResponse(w http.ResponseWriter, msg string, status int) {
	json.NewEncoder(w).Encode(&errResponse{
		message: msg,
		status:  status,
	})
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Accept", "application/json")
	})
}

func (app *app) createItem(w http.ResponseWriter, r *http.Request) {
	coll := app.mc.Database(os.Getenv("DBNAME")).Collection(os.Getenv("DBCOLL"))
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		serveErrResponse(w, "bad request", http.StatusBadRequest)
		return
	}

	item, err := model.UnmarshalItem(bodyBytes)
	if err != nil {
		serveErrResponse(w, "err unmarshaling", http.StatusBadRequest)
		return
	}

	res, err := db.InsertOneItem(r.Context(), coll, &item)
	if err != nil {
		serveErrResponse(w, "err unmarshaling", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(&res)
}

func (app *app) listSingleItem(w http.ResponseWriter, r *http.Request) {

}

func (app *app) listItems(w http.ResponseWriter, r *http.Request) {

}

func (app *app) updateItem(w http.ResponseWriter, r *http.Request) {

}

func (app *app) deleteItem(w http.ResponseWriter, r *http.Request) {

}

func CreateRouter(app *app) *mux.Router {
	r := mux.NewRouter()

	r.Use(commonMiddleware)

	i := r.PathPrefix("/items").Subrouter()
	i.HandleFunc("/create", app.createItem).Methods(http.MethodPost)
	i.HandleFunc("/list/{id}", app.listSingleItem).Methods(http.MethodGet)
	i.HandleFunc("/list", app.listItems).Methods(http.MethodGet)
	i.HandleFunc("/update/{id}", app.updateItem).Methods(http.MethodPut)
	i.HandleFunc("/delete/{id}", app.deleteItem).Methods(http.MethodDelete)

	return r
}

func CreateServer() (*http.Server, error) {
	app, err := CreateApp()
	router := CreateRouter(app)

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", os.Getenv("SERVERPORT")),
		Handler: router,
	}, err
}
