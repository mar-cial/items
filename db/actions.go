package db

import (
	"context"

	"github.com/mar-cial/items/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func InsertOneItem(ctx context.Context, coll *mongo.Collection, item *model.Item) (*mongo.InsertOneResult, error) {
	bsonDoc, err := bson.Marshal(item)
	if err != nil {
		return &mongo.InsertOneResult{}, err
	}

	return coll.InsertOne(ctx, bsonDoc)
}

func InsertItems(ctx context.Context, coll *mongo.Collection, items []model.Item) (*mongo.InsertManyResult, error) {
	var in []interface{}

	for k := range items {
		in = append(in, items[k])
	}
	return coll.InsertMany(ctx, in)
}

func ListOneItem(ctx context.Context, coll *mongo.Collection, id string) (model.Item, error) {
	var err error
	var result model.Item

	mongoid, err := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": mongoid}

	err = coll.FindOne(ctx, filter).Decode(&result)
	return result, err
}

func ListItems(ctx context.Context, coll *mongo.Collection) ([]model.Item, error) {
	var err error
	var results []model.Item

	// empty bson.M{} means "find everything, no filter". Just leaving it here in
	// case it needs to change later.
	filter := bson.M{}

	cursor, err := coll.Find(ctx, filter)

	err = cursor.All(ctx, &results)

	return results, err
}

func UpdateOneItem(ctx context.Context, coll *mongo.Collection, id string, item *model.Item) (*mongo.UpdateResult, error) {
	var err error
	mongoid, err := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": mongoid}
	update := bson.M{"$set": bson.M{"title": item.Title, "price": item.Price}}

	res, err := coll.UpdateOne(ctx, filter, update)
	return res, err
}

func DeleteOneItem(ctx context.Context, coll *mongo.Collection, id string) (*mongo.DeleteResult, error) {
	mongoid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return &mongo.DeleteResult{}, err
	}

	filter := bson.M{"_id": mongoid}

	return coll.DeleteOne(ctx, filter)
}
