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

func (app *app) createOneItemHandler(w http.ResponseWriter, r *http.Request) {
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

func (app *app) createManyItemsHandler(w http.ResponseWriter, r *http.Request) {
	coll := app.mc.Database(os.Getenv("DBNAME")).Collection(os.Getenv("DBCOLL"))

	var items []model.Item
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		serveErrResponse(w, "err unmarshaling", http.StatusBadRequest)
		return
	}

	insertManyRes, err := db.InsertItems(r.Context(), coll, items)
	if err != nil {
		serveErrResponse(w, "err from inserting items", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(&insertManyRes)

}

func (app *app) listOneItemHandler(w http.ResponseWriter, r *http.Request) {
	coll := app.mc.Database(os.Getenv("DBNAME")).Collection(os.Getenv("DBCOLL"))

	id := mux.Vars(r)["id"]
	if id == "" {
		serveErrResponse(w, "no id received", http.StatusInternalServerError)
		return
	}

	item, err := db.ListOneItem(r.Context(), coll, id)
	if err != nil {
		serveErrResponse(w, "err listing an item", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(&item)
}

func (app *app) listItemsHandler(w http.ResponseWriter, r *http.Request) {
	coll := app.mc.Database(os.Getenv("DBNAME")).Collection(os.Getenv("DBCOLL"))

	items, err := db.ListItems(r.Context(), coll)
	if err != nil {
		serveErrResponse(w, "could not list all items", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(&items)
}

func (app *app) updateOneItemHandler(w http.ResponseWriter, r *http.Request) {

}

func (app *app) deleteOneItemHandler(w http.ResponseWriter, r *http.Request) {

}

func CreateRouter(app *app) *mux.Router {
	r := mux.NewRouter()

	r.Use(commonMiddleware)

	// an httprouter kinda approach...
	// I'm not familiar with httprouter so I'll just use gorilla mux
	i := r.PathPrefix("/items").Subrouter()
	i.HandleFunc("/create/one", app.createOneItemHandler).Methods(http.MethodPost)
	i.HandleFunc("/create/many", app.createManyItemsHandler).Methods(http.MethodPost)
	i.HandleFunc("/list/{id}", app.listOneItemHandler).Methods(http.MethodGet)
	i.HandleFunc("/list", app.listItemsHandler).Methods(http.MethodGet)
	i.HandleFunc("/update/{id}", app.updateOneItemHandler).Methods(http.MethodPut)
	i.HandleFunc("/delete/{id}", app.deleteOneItemHandler).Methods(http.MethodDelete)

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
