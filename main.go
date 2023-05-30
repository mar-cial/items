package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	godotenv.Load()
}

type Item struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Title string             `json:"title" bson:"title"`
	Price float64            `json:"price" bson:"price"`
}

type ErrResponse struct {
	Status  int    `json:"Status"`
	Message string `json:"Message"`
}

func insertItems(ctx context.Context, coll *mongo.Collection, items []Item) (*mongo.InsertManyResult, error) {

	var itemsToInsert []interface{}

	for _, v := range items {
		itemsToInsert = append(itemsToInsert, v)
	}

	return coll.InsertMany(ctx, itemsToInsert)
}

func listItems(w http.ResponseWriter, r *http.Request) {

}

type app struct {
	coll *mongo.Collection
}

func createClient() *mongo.Client {
	uri := os.Getenv("MONGO_URI")
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		log.Fatalln("err creating client: ", err)
	}

	return client
}

func NewApp() *app {
	coll := createClient().Database("DB_NAME").Collection("DB_COLL")

	return &app{
		coll: coll,
	}
}

func (app *app) createItems(w http.ResponseWriter, r *http.Request) {
	var items []Item
	err := json.NewDecoder(r.Body).Decode(&items)

	if err != nil {

		errRes := ErrResponse{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(errRes)
	}

	res, err := insertItems(context.Background(), app.coll, items)
	if err != nil {
		errRes := ErrResponse{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(errRes)
	}

	json.NewEncoder(w).Encode(res)
}

func router(app *app) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/create", app.createItems).Methods(http.MethodPost)

	return r
}

func main() {
	app := NewApp()
	r := router(app)

	server := http.Server{
		Addr:    ":8000",
		Handler: r,
	}

	log.Fatalln(server.ListenAndServe())
}
