package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/mar-cial/items/model"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	endpoint string
	mc       *mongo.Client
	a        *app
	ids      []string
)

func TestCreateApp(t *testing.T) {
	app, err := CreateApp()
	assert.NoError(t, err)

	a = app
	mc = app.mc

	err = mc.Ping(context.Background(), nil)
	assert.NoError(t, err)
}

func TestCreateServer(t *testing.T) {
	server, err := CreateServer()
	assert.NoError(t, err)
	assert.Equal(t, ":8000", server.Addr)
}

func TestCreateRouter(t *testing.T) {
	// r := CreateRouter(a)
	// will come back to this
}

// I guess I'll give it this long ass name with sufix -Handler to
// REALLY differentiate them
func TestCreateSingleItemHandler(t *testing.T) {
	testItem := model.Item{
		Title: "Test Item",
		Price: 69.69,
	}

	itemBytes, err := testItem.Marshal()
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/items/create/one", bytes.NewReader(itemBytes))
	w := httptest.NewRecorder()

	a.createOneItemHandler(w, req)

	res := w.Result()

	defer func() {
		err := res.Body.Close()
		log.Fatal(err)
	}()

	// assertions

	decoder := json.NewDecoder(res.Body)

	var insertRes *mongo.InsertOneResult

	err = decoder.Decode(&insertRes)
	assert.NoError(t, err)
	assert.True(t, primitive.IsValidObjectID(insertRes.InsertedID.(string)))

	ids = append(ids, insertRes.InsertedID.(string))
}

func TestCreateItemsHandler(t *testing.T) {
	testItems := []model.Item{
		{
			Title: "Test item 2",
			Price: 12.24,
		},
		{
			Title: "Test item 3",
			Price: 52.24,
		},
	}

	itemsBytes, err := json.Marshal(testItems)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/items/create/many", bytes.NewReader(itemsBytes))
	w := httptest.NewRecorder()

	a.createManyItemsHandler(w, req)

	res := w.Result()
	defer func() {
		err := res.Body.Close()
		log.Fatal(err)
	}()

	decoder := json.NewDecoder(res.Body)

	var insertManyRes *mongo.InsertManyResult

	err = decoder.Decode(&insertManyRes)
	assert.NoError(t, err)

	for a := range insertManyRes.InsertedIDs {
		assert.True(t, primitive.IsValidObjectID(insertManyRes.InsertedIDs[a].(string)))
		ids = append(ids, insertManyRes.InsertedIDs[a].(string))
	}
}

func TestListOneItemHandler(t *testing.T) {
	path := fmt.Sprintf("/items/%s", ids[0])
	assert.True(t, primitive.IsValidObjectID(ids[0]))

	req := httptest.NewRequest(http.MethodGet, path, nil)

	rec := httptest.NewRecorder()

	// have to create a new mux router in order to process that id var
	router := mux.NewRouter()

	router.HandleFunc("/items/{id}", a.listOneItemHandler)

	router.ServeHTTP(rec, req)

	res := rec.Result()
	defer func() {
		err := res.Body.Close()
		log.Fatal(err)
	}()

	decoder := json.NewDecoder(res.Body)

	var item *model.Item
	err := decoder.Decode(&item)
	assert.NoError(t, err)

	// assertions
	assert.True(t, primitive.IsValidObjectID(item.ID.Hex()))
	assert.NotEmpty(t, item.Title)
	assert.Greater(t, item.Price, 0.0)
}

func TestListItemsHandler(t *testing.T) {
	path := "/items/list"

	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()

	router := mux.NewRouter()

	router.HandleFunc(path, a.listItemsHandler)

	router.ServeHTTP(rec, req)

	res := rec.Result()
	defer func() {
		err := res.Body.Close()
		log.Fatal(err)
	}()

	decoder := json.NewDecoder(res.Body)

	var items []model.Item
	err := decoder.Decode(&items)
	assert.NoError(t, err)

	// assertions
	for a := range items {
		assert.True(t, primitive.IsValidObjectID(items[a].ID.Hex()))
		assert.NotEmpty(t, items[a].Title)
		assert.Greater(t, items[a].Price, 0.0)
	}
}

func TestUpdateItemsHandler(t *testing.T) {
	path := fmt.Sprintf("/items/update/%s", ids[0])

	assert.True(t, primitive.IsValidObjectID(ids[0]))

	updatedItem := &model.Item{
		Title: "Updated item",
		Price: 8008.50,
	}

	itemBytes, err := json.Marshal(updatedItem)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(itemBytes))
	rec := httptest.NewRecorder()

	// have to create a new mux router in order to process that id var
	router := mux.NewRouter()

	router.HandleFunc("/items/update/{id}", a.updateOneItemHandler)

	router.ServeHTTP(rec, req)

	res := rec.Result()
	defer func() {
		err := res.Body.Close()
		log.Fatal(err)
	}()

	decoder := json.NewDecoder(res.Body)

	var updateResult *mongo.UpdateResult
	err = decoder.Decode(&updateResult)
	assert.NoError(t, err)

	assert.Equal(t, int64(1), updateResult.MatchedCount)
	assert.Equal(t, int64(1), updateResult.ModifiedCount)
}

func TestDeleteOneItemHandler(t *testing.T) {
	path := fmt.Sprintf("/items/delete/%s", ids[0])
	assert.True(t, primitive.IsValidObjectID(ids[0]))

	req := httptest.NewRequest(http.MethodDelete, path, nil)
	rec := httptest.NewRecorder()

	// another router... maybe I should've used that router up there...
	router := mux.NewRouter()

	router.HandleFunc("/items/delete/{id}", a.deleteOneItemHandler).Methods(http.MethodDelete)

	router.ServeHTTP(rec, req)

	res := rec.Result()
	defer func() {
		err := res.Body.Close()
		log.Fatal(err)
	}()

	decoder := json.NewDecoder(res.Body)

	var deleteResult *mongo.DeleteResult
	err := decoder.Decode(&deleteResult)
	assert.NoError(t, err)

	assert.Equal(t, int64(1), deleteResult.DeletedCount)
}

func TestCheckItemsAgain(t *testing.T) {
	path := "/items/list"

	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()

	router := mux.NewRouter()

	router.HandleFunc(path, a.listItemsHandler)

	router.ServeHTTP(rec, req)

	res := rec.Result()
	defer func() {
		err := res.Body.Close()
		log.Fatal(err)
	}()

	decoder := json.NewDecoder(res.Body)

	var items []model.Item
	err := decoder.Decode(&items)
	assert.NoError(t, err)

	// assertions
	for a := range items {
		assert.True(t, primitive.IsValidObjectID(items[a].ID.Hex()))
		assert.NotEmpty(t, items[a].Title)
		assert.Greater(t, items[a].Price, 0.0)
	}
}

func CreateMongoContainer() (testcontainers.Container, error) {
	envmap := map[string]string{
		"DBUSER":     "root",
		"DBPASS":     "testpass",
		"DBPORT":     "27017/tcp",
		"DBNAME":     "testdb",
		"DBCOLL":     "testcoll",
		"SERVERPORT": "8000",
	}

	for k := range envmap {
		err := os.Setenv(k, envmap[k])
		if err != nil {
			log.Fatal(err)
		}
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
		Name:         "apiPkgMongoTestContainer",
		Hostname:     os.Getenv("DBHOST"),
		AutoRemove:   true,
	}

	mongoC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatal(err)
	}

	return mongoC, err
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	mongoC, err := CreateMongoContainer()
	if err != nil {
		log.Fatal(err)
	}

	endpoint, err := mongoC.Endpoint(ctx, "")
	if err != nil {
		log.Fatal(err)
	}

	err = os.Setenv("MONGOURI", fmt.Sprintf("mongodb://%s:%s@%s", os.Getenv("DBUSER"), os.Getenv("DBPASS"), endpoint))
	if err != nil {
		log.Fatal(err)
	}

	m.Run()

	defer func() {
		os.Clearenv()
		err := mongoC.Terminate(ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}()
}
