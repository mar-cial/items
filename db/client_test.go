package db

import (
	"context"
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
var c *mongo.Client
var testItem *model.Item
var testItemsList []model.Item

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

	c = client
}

func TestInsertItems(t *testing.T) {
	ctx := context.Background()

	items := []model.Item{
		{
			ID:    primitive.NewObjectID(),
			Title: "Test Product 1",
			Price: 420.69,
		},
		{
			ID:    primitive.NewObjectID(),
			Title: "Test Product 2",
			Price: 99.99,
		},
	}

	// we are going to use this item to list, update and delete it
	testItem = &items[0]

	dbname := os.Getenv("DBNAME")
	dbcoll := os.Getenv("DBCOLL")

	coll := c.Database(dbname).Collection(dbcoll)

	res, err := InsertItems(ctx, coll, items)
	assert.NoError(t, err)

	assert.Equal(t, items[0].ID, res.InsertedIDs[0])
	assert.Len(t, res.InsertedIDs, 2)
}

func TestListSingleItem(t *testing.T) {
	ctx := context.Background()

	dbname := os.Getenv("DBNAME")
	dbcoll := os.Getenv("DBCOLL")

	coll := c.Database(dbname).Collection(dbcoll)

	res, err := ListSingleItem(ctx, coll, testItem.ID.Hex())
	assert.NoError(t, err)

	assert.Equal(t, testItem.ID, res.ID)
	assert.Equal(t, testItem.Title, res.Title)
	assert.Equal(t, testItem.Price, res.Price)
}

func TestListItems(t *testing.T) {
	ctx := context.Background()

	dbname := os.Getenv("DBNAME")
	dbcoll := os.Getenv("DBCOLL")

	coll := c.Database(dbname).Collection(dbcoll)

	res, err := ListItems(ctx, coll)
	assert.NoError(t, err)

	for i, v := range testItemsList {

		assert.Equal(t, v.ID, res[i].ID, "should be same")

	}
}

func TestUpdateOneItem(t *testing.T) {
	ctx := context.Background()

	dbname := os.Getenv("DBNAME")
	dbcoll := os.Getenv("DBCOLL")

	coll := c.Database(dbname).Collection(dbcoll)

	newItem := model.Item{
		Title: "Updated item",
		Price: 12.99,
	}

	res, err := UpdateOneItem(ctx, coll, testItem.ID.Hex(), &newItem)
	assert.NoError(t, err)

	// needs to be converted to int64 to pass, for whatever reason
	assert.Equal(t, int64(1), res.MatchedCount)
	assert.Equal(t, int64(1), res.ModifiedCount)

	// get the item that was just inserted
	item, err := ListSingleItem(ctx, coll, testItem.ID.Hex())
	assert.NoError(t, err)

	// make sure values match
	assert.Equal(t, newItem.Title, item.Title)
	assert.Equal(t, newItem.Price, item.Price)
}

func TestDeleteOneItem(t *testing.T) {
	ctx := context.Background()

	dbname := os.Getenv("DBNAME")
	dbcoll := os.Getenv("DBCOLL")

	coll := c.Database(dbname).Collection(dbcoll)

	item, err := ListSingleItem(ctx, coll, testItem.ID.Hex())
	assert.NoError(t, err)

	res, err := DeleteOneItem(ctx, coll, item.ID.Hex())
	assert.NoError(t, err)

	assert.Equal(t, int64(1), res.DeletedCount)
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
		Name:         "mongoTestContainer",
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
