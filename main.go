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
	coll := createClient().Database(os.Getenv("DB_NAME")).Collection(os.Getenv("DB_COLL"))
	return &app{
		coll: coll,
	}
}

func (app *app) createItems(w http.ResponseWriter, r *http.Request) {
	var items []Item
	err := json.NewDecoder(r.Body).Decode(&items)
	if err != nil {
		errResponse := ErrResponse{
			Status:  http.StatusBadRequest,
			Message: "Err decoding items",
		}

		json.NewEncoder(w).Encode(errResponse)
		return
	}

	for _, v := range items {
		if len(v.Title) < 3 {
			errResponse := ErrResponse{
				Status:  http.StatusBadRequest,
				Message: "Title must be at least 3 chars long. Either title is too short or wrong type",
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		if v.Price <= 0 {
			errResponse := ErrResponse{
				Status:  http.StatusBadRequest,
				Message: "Price cannot be zero, nil or something else happened",
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}
	}

	res, err := insertItems(context.Background(), app.coll, items)
	if err != nil {
		errResponse := ErrResponse{
			Status:  http.StatusInternalServerError,
			Message: "Unable to insert items",
		}

		json.NewEncoder(w).Encode(errResponse)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (app *app) listItems(w http.ResponseWriter, r *http.Request) {

	items, err := getItems(context.Background(), app.coll)
	if err != nil {
		if len(items) == 0 || items == nil {
			errResponse := ErrResponse{
				Status:  http.StatusBadRequest,
				Message: "items are either nil or len(items) == 0",
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		} else {
			errResponse := ErrResponse{
				Status:  http.StatusInternalServerError,
				Message: "Something went wrong while getting items",
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

	}
	w.WriteHeader(http.StatusOK)
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
	if err != nil {
		errResponse := ErrResponse{
			Status:  http.StatusInternalServerError,
			Message: "unable to generate ObjectID from the provided id",
		}

		json.NewEncoder(w).Encode(errResponse)
		return

	}

	item, err := getItem(context.Background(), app.coll, mongoid)
	if err != nil {
		errResponse := ErrResponse{
			Status:  http.StatusInternalServerError,
			Message: "unable to get single item",
		}

		json.NewEncoder(w).Encode(errResponse)
		return

	}

	w.WriteHeader(http.StatusOK)
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
	if err != nil {
		if len(item.Title) < 3 {
			errResponse := ErrResponse{
				Status:  http.StatusBadRequest,
				Message: "Title must be at least 3 chars long. Either title is too short or",
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		if item.Price <= 0 {
			errResponse := ErrResponse{
				Status:  http.StatusBadRequest,
				Message: "Price cannot be zero, nil or something else happened",
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		errResponse := ErrResponse{
			Status:  http.StatusUnprocessableEntity,
			Message: "Cannot process request",
		}

		json.NewEncoder(w).Encode(errResponse)
		return

	}

	res, err := updateItem(context.Background(), app.coll, id, item)
	if err != nil {
		errResponse := ErrResponse{
			Status:  http.StatusInternalServerError,
			Message: "Something went wrong while updating item",
		}

		json.NewEncoder(w).Encode(errResponse)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func deleteItem(ctx context.Context, coll *mongo.Collection, id string) (*mongo.DeleteResult, error) {
	mongoid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return &mongo.DeleteResult{}, err
	}

	return coll.DeleteOne(ctx, bson.M{"_id": mongoid})
}

func (app *app) deleteSingleItem(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if len(id) != 24 {
		errResponse := ErrResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid ID string length. Should be 24.",
		}

		json.NewEncoder(w).Encode(errResponse)
		return

	}

	res, err := deleteItem(context.Background(), app.coll, id)
	if err != nil {
		errResponse := ErrResponse{
			Status:  http.StatusInternalServerError,
			Message: "Something went wrong while deleting item",
		}

		json.NewEncoder(w).Encode(errResponse)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func router(app *app) *mux.Router {
	r := mux.NewRouter()

	r.Use(commonMiddleware)

	r.HandleFunc("/items", app.createItems).Methods(http.MethodPost)
	r.HandleFunc("/items", app.listItems).Methods(http.MethodGet)
	r.HandleFunc("/items/{id}", app.listSingleItem).Methods(http.MethodGet)
	r.HandleFunc("/items/{id}", app.updateSingleItem).Methods(http.MethodPut)
	r.HandleFunc("/items/{id}", app.deleteSingleItem).Methods(http.MethodDelete)

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
