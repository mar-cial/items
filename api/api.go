package api

import (
	"encoding/json"
	"errors"
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
	client, err := db.CreateClient(os.Getenv("MONGOURI"))

	return &app{
		mc: client,
	}, err
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	item, err := model.UnmarshalItem(bodyBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := db.InsertOneItem(r.Context(), coll, &item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *app) createManyItemsHandler(w http.ResponseWriter, r *http.Request) {
	coll := app.mc.Database(os.Getenv("DBNAME")).Collection(os.Getenv("DBCOLL"))

	var items []model.Item
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	insertManyRes, err := db.InsertItems(r.Context(), coll, items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&insertManyRes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *app) listOneItemHandler(w http.ResponseWriter, r *http.Request) {
	coll := app.mc.Database(os.Getenv("DBNAME")).Collection(os.Getenv("DBCOLL"))

	id := mux.Vars(r)["id"]
	if id == "" {
		err := errors.New("no id provided")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	item, err := db.ListOneItem(r.Context(), coll, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&item); err != nil {

	}
}

func (app *app) listItemsHandler(w http.ResponseWriter, r *http.Request) {
	coll := app.mc.Database(os.Getenv("DBNAME")).Collection(os.Getenv("DBCOLL"))

	items, err := db.ListItems(r.Context(), coll)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *app) updateOneItemHandler(w http.ResponseWriter, r *http.Request) {
	coll := app.mc.Database(os.Getenv("DBNAME")).Collection(os.Getenv("DBCOLL"))
	id := mux.Vars(r)["id"]

	if id == "" {
		err := errors.New("no id provided")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var item *model.Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updateRes, err := db.UpdateOneItem(r.Context(), coll, id, item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(updateRes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (app *app) deleteOneItemHandler(w http.ResponseWriter, r *http.Request) {
	coll := app.mc.Database(os.Getenv("DBNAME")).Collection(os.Getenv("DBCOLL"))
	id := mux.Vars(r)["id"]

	if id == "" {
		err := errors.New("no id provided")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	delRes, err := db.DeleteOneItem(r.Context(), coll, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.NewEncoder(w).Encode(delRes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
