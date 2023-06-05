package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Item struct {
	ID    primitive.ObjectID `bson:"_id, omitempty"`
	Title string             `json:"title" bson:"title"`
	Price float64            `json:"price" bson:"price"`
}
