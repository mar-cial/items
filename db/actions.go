package db

import (
	"context"

	"github.com/mar-cial/items/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func InsertItems(ctx context.Context, coll *mongo.Collection, items []model.Item) (*mongo.InsertManyResult, error) {
	var in []interface{}

	for k := range items {
		in = append(in, items[k])
	}

	return coll.InsertMany(ctx, in)
}

func ListSingleItem(ctx context.Context, coll *mongo.Collection, id string) (model.Item, error) {
	var err error
	var result model.Item

	mongoid, err := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": mongoid}

	err = coll.FindOne(ctx, filter).Decode(&result)
	return result, err
}
