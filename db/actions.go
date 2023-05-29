package db

import (
	"context"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/mar-cial/items/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)
)

func createClient() (*mongo.Client, error) {
	uri := os.Getenv("MONGO_URI")
	opts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatalln("err connecting to mongodb: ", err)
	}

	return client, err
}

// db instance
type DB struct {
	client *mongo.Client
}

func NewDBInstance() *DB {
	c, err := createClient()
	if err != nil {
		log.Fatalln("could not create client")
	}

	return &DB{
		client: c,
	}
}

func (db *DB) GetItems() ([]model.Item, error) {
	coll := db.client.Database("itemsdb").Collection("items")

	var items []model.Item
	cur, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return items, err
	}

	err = cur.All(ctx, &items)
	if err != nil {
		return items, err
	}

	return items, err
}

func (db *DB) CreateItems() (*mongo.InsertManyResult, error) {
	coll := db.client.Database("itemsdb").Collection("items")

	newItems := []interface{}{
		model.Item{
			ID:    primitive.NewObjectID(),
			Title: "Test Product",
			Price: 100 + rand.Float64()*1000,
		},
		model.Item{
			ID:    primitive.NewObjectID(),
			Title: "Test Product",
			Price: 100 + rand.Float64()*1000,
		},
		model.Item{
			ID:    primitive.NewObjectID(),
			Title: "Test Product",
			Price: 100 + rand.Float64()*1000,
		},
		model.Item{
			ID:    primitive.NewObjectID(),
			Title: "Test Product",
			Price: 100 + rand.Float64()*1000,
		},
	}

	result, err := coll.InsertMany(ctx, newItems)
	return result, err
}
