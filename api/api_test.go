package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mar-cial/items/model"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"go.mongodb.org/mongo-driver/mongo"
)

var endpoint string
var mc *mongo.Client
var testItem *model.Item
var testItemsList []model.Item
var a *app

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

func TestCreateSingleItem(t *testing.T) {
	testItem := model.Item{
		Title: "Test Item",
		Price: 69.69,
	}

	itemBytes, err := testItem.Marshal()
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/items/list", bytes.NewBuffer(itemBytes))
	w := httptest.NewRecorder()

	a.createItem(w, req)

	res := w.Result()
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	assert.NoError(t, err)

	dataString := string(data)
	assert.Contains(t, dataString, "InsertedID")

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
