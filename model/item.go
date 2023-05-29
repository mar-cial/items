package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Item struct {
	ID    primitive.ObjectID `json:"_id" bson:"_id"`
	Title string             `json:"title" bson:"title"`
	Price float64            `json:"price" bson:"price"`
}
