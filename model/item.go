package model

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Item struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Title string             `json:"title" bson:"title"`
	Price float64            `json:"price" bson:"price"`
}

func UnmarshalItem(data []byte) (Item, error) {
	var r Item
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Item) Marshal() ([]byte, error) {
	return json.Marshal(r)
}
