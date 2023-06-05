package db

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateClient() (*mongo.Client, error) {
	user := os.Getenv("DBUSER")
	pass := os.Getenv("DBPASS")
	host := os.Getenv("DBHOST")
	port := os.Getenv("DBPORT")

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s", user, pass, host, port)
	return mongo.Connect(
		context.Background(),
		options.Client().ApplyURI(uri))

}
