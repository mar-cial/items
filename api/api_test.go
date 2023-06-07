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

var endpoint string
var mc *mongo.Client
var testItem *model.Item
var testItemsList []model.Item
var a *app
var ids []string

func TestCreateApp(t *testing.T) {
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
	router := CreateRouter(a)
	// will come back to this
	fmt.Println(&router)
}

// I guess I'll give it this long ass name with sufix -Handler to REALLY differentiate them
// db/actions
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
	defer res.Body.Close()

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
	defer res.Body.Close()

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

	decoder := json.NewDecoder(rec.Result().Body)

	var item model.Item
	err := decoder.Decode(&item)
	if err != nil {
		fmt.Println("err: ", err)
	}

	// assertions
	assert.True(t, primitive.IsValidObjectID(item.ID.Hex()))
	assert.NotEmpty(t, item.Title)
	assert.Greater(t, item.Price, 0.0)
}

func TestListItems(t *testing.T) {
	// path := "/items/list"

	// need to check every item returned by the function
	// req := httptest.NewRequest(http.MethodGet, path, nil)
	// rec := httptest.NewRecoder()

}

func TestMain(m *testing.M) {
	var err error
	envmap := map[string]string{
		"DBUSER":     "root",
		"DBPASS":     "testpass",
		"DBHOST":     "localhost",
		"DBPORT":     "27017",
		"DBNAME":     "testdb",
		"DBCOLL":     "testcoll",
		"SERVERPORT": "8000",
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
		Name:         "apiPkgMongoTestContainer",
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
		var err error
		os.Clearenv()
		err = mongoC.Terminate(ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}()
}
