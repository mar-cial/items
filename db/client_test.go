package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/mar-cial/items/model"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var endpoint string
var mc *mongo.Client
var ids []primitive.ObjectID

func TestCreateClient(t *testing.T) {
	var err error
	arrs := []string{"DBUSER", "DBPASS", "DBHOST", "DBPORT", "DBNAME"}
	for k := range arrs {
		s := os.Getenv(arrs[k])
		assert.NotEmpty(t, s)
	}
	assert.NoError(t, err)

	endpoint = endpoint[len(endpoint)-5:]
	err = os.Setenv("DBPORT", endpoint)
	assert.NoError(t, err)

	client, err := CreateClient()
	assert.NotEmpty(t, client)
	assert.NoError(t, err)

	err = client.Ping(context.Background(), nil)
	assert.NoError(t, err)

	mc = client
}

func TestInsertItem(t *testing.T) {
	ctx := context.Background()

	item := model.Item{
		Title: "Test Product 1",
		Price: 8008.55,
	}

	dbname := os.Getenv("DBNAME")
	dbcoll := os.Getenv("DBCOLL")

	coll := mc.Database(dbname).Collection(dbcoll)

	res, err := InsertOneItem(ctx, coll, &item)
	assert.NoError(t, err)

	assert.True(t, primitive.IsValidObjectID(res.InsertedID.(primitive.ObjectID).Hex()))

	ids = append(ids, res.InsertedID.(primitive.ObjectID))
}

func TestInsertItems(t *testing.T) {
	ctx := context.Background()

	// we are also going to append this slice to testItemsList
	items := []model.Item{
		{
			Title: "Test Product 2",
			Price: 420.69,
		},
		{
			Title: "Test Product 3",
			Price: 99.99,
		},
	}

	dbname := os.Getenv("DBNAME")
	dbcoll := os.Getenv("DBCOLL")

	coll := mc.Database(dbname).Collection(dbcoll)

	res, err := InsertItems(ctx, coll, items)
	assert.NoError(t, err)

	assert.Len(t, res.InsertedIDs, 2)
	for k := range res.InsertedIDs {
		assert.True(t, primitive.IsValidObjectID(res.InsertedIDs[k].(primitive.ObjectID).Hex()))
		ids = append(ids, res.InsertedIDs[k].(primitive.ObjectID))
	}

}

func TestListOneItem(t *testing.T) {
	ctx := context.Background()

	dbname := os.Getenv("DBNAME")
	dbcoll := os.Getenv("DBCOLL")

	coll := mc.Database(dbname).Collection(dbcoll)

	res, err := ListOneItem(ctx, coll, ids[0].Hex())
	assert.NoError(t, err)

	assert.True(t, primitive.IsValidObjectID(res.ID.Hex()))
	assert.Greater(t, res.Price, 0.0)
	assert.NotEmpty(t, res.Title)
}

func TestListItems(t *testing.T) {
	ctx := context.Background()

	dbname := os.Getenv("DBNAME")
	dbcoll := os.Getenv("DBCOLL")

	coll := mc.Database(dbname).Collection(dbcoll)

	res, err := ListItems(ctx, coll)
	assert.NoError(t, err)

	for k := range res {
		assert.True(t, primitive.IsValidObjectID(res[k].ID.Hex()))
		assert.Greater(t, res[k].Price, 0.0)
		assert.NotEmpty(t, res[k].Title)
	}
}

func TestUpdateOneItem(t *testing.T) {
	ctx := context.Background()

	dbname := os.Getenv("DBNAME")
	dbcoll := os.Getenv("DBCOLL")

	coll := mc.Database(dbname).Collection(dbcoll)

	updatedItem := &model.Item{
		Title: "Updated Item 1",
		Price: 80085.55,
	}

	res, err := UpdateOneItem(ctx, coll, ids[0].Hex(), updatedItem)
	assert.NoError(t, err)

	// needs to be converted to int64 to pass, for whatever reason
	assert.Equal(t, int64(1), res.MatchedCount)
	assert.Equal(t, int64(1), res.ModifiedCount)

	// checking if it updated
	item, _ := ListOneItem(ctx, coll, ids[0].Hex())

	assert.Equal(t, item.Title, updatedItem.Title)
	assert.Equal(t, item.Price, updatedItem.Price)
}

func TestDeleteOneItem(t *testing.T) {
	ctx := context.Background()

	dbname := os.Getenv("DBNAME")
	dbcoll := os.Getenv("DBCOLL")

	coll := mc.Database(dbname).Collection(dbcoll)

	res, err := DeleteOneItem(ctx, coll, ids[0].Hex())
	assert.NoError(t, err)

	fmt.Println(res)
}

func TestMain(m *testing.M) {
	var err error
	envmap := map[string]string{
		"DBUSER": "root",
		"DBPASS": "testpass",
		"DBHOST": "localhost",
		"DBPORT": "27017",
		"DBNAME": "testdb",
		"DBCOLL": "testcoll",
	}

	for k := range envmap {
		err = os.Setenv(k, envmap[k])
	}

	// to do this we need a mongo image, and a container running that image.
	// aight the container is running now.
	envs := map[string]string{
		"MONGO_INITDB_ROOT_USERNAME": os.Getenv("DBUSER"),
		"MONGO_INITDB_ROOT_PASSWORD": os.Getenv("DBPASS"),
		"MONGO_INITDB_DATABASE":      os.Getenv("DBNAME"),
	}

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "mongo",
		Env:          envs,
		ExposedPorts: []string{os.Getenv("DBPORT")},
		Name:         "dbPkgMongoTestContainer",
		Hostname:     os.Getenv("DBHOST"),
		AutoRemove:   true,
	}

	mongoC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	endpoint, err = mongoC.Endpoint(ctx, "")
	if err != nil {
		log.Fatalln(err)
	}

	m.Run()

	defer func() {
		os.Clearenv()
		err = mongoC.Terminate(ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}()
}
