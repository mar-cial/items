package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
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

func EncodeError(w http.ResponseWriter, err error) {
	if err != nil {
		errRes := ErrResponse{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(errRes)
		return
	}
}

func insertItems(ctx context.Context, coll *mongo.Collection, items []Item) (*mongo.InsertManyResult, error) {

	var itemsToInsert []interface{}

	for _, v := range items {
		itemsToInsert = append(itemsToInsert, v)
	}

	return coll.InsertMany(ctx, itemsToInsert)
}

func getItems(ctx context.Context, coll *mongo.Collection) ([]Item, error) {
	filter := bson.M{}

	var items []Item
	cursor, err := coll.Find(context.Background(), filter)
	if err != nil {
		return items, err
	}

	err = cursor.All(ctx, &items)
	return items, err
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
	EncodeError(w, err)

	res, err := insertItems(context.Background(), app.coll, items)
	EncodeError(w, err)

	json.NewEncoder(w).Encode(res)
}

func (app *app) listItems(w http.ResponseWriter, r *http.Request) {
	items, err := getItems(context.Background(), app.coll)
	EncodeError(w, err)

	json.NewEncoder(w).Encode(items)
}

func getItem(ctx context.Context, coll *mongo.Collection, id primitive.ObjectID) (Item, error) {
	var result Item
	filter := bson.M{"_id": id}
	err := coll.FindOne(context.Background(), filter).Decode(&result)
	return result, err
}

func (app *app) listSingleItem(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	mongoid, err := primitive.ObjectIDFromHex(id)
	EncodeError(w, err)

	item, err := getItem(context.Background(), app.coll, mongoid)
	EncodeError(w, err)

	json.NewEncoder(w).Encode(item)
}

func updateItem(ctx context.Context, coll *mongo.Collection, id string, in Item) (*mongo.UpdateResult, error) {

	mongoid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return &mongo.UpdateResult{}, err
	}

	updateItem := bson.M{"title": in.Title, "price": in.Price}

	return coll.UpdateByID(ctx, mongoid, bson.M{"$set": updateItem})
}

func (app *app) updateSingleItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	id := mux.Vars(r)["id"]

	err := json.NewDecoder(r.Body).Decode(&item)
	EncodeError(w, err)

	res, err := updateItem(context.Background(), app.coll, id, item)
	EncodeError(w, err)

	json.NewEncoder(w).Encode(res)
}

func router(app *app) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/items", app.createItems).Methods(http.MethodPost)
	r.HandleFunc("/items", app.listItems).Methods(http.MethodGet)
	r.HandleFunc("/items/{id}", app.listSingleItem).Methods(http.MethodGet)
	r.HandleFunc("/items/{id}", app.updateSingleItem).Methods(http.MethodPut)

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
