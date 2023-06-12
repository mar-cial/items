package db

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateClient(uri string) (*mongo.Client, error) {
	if len(uri) > 0 {
		return mongo.Connect(
			context.Background(),
			options.Client().ApplyURI(uri))

	}

	err := errors.New("could not connect")
	fmt.Println(err)
	return nil, err
}
