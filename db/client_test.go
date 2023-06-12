package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	endpoint string
	ids      []primitive.ObjectID
)

func TestCreateClient(t *testing.T) {
	arrs := []string{"DBUSER", "DBPASS", "DBHOST", "DBPORT", "DBNAME"}
	for k := range arrs {
		s := os.Getenv(arrs[k])
		assert.NotEmpty(t, s)
	}

	endpoint = endpoint[len(endpoint)-5:]
	err := os.Setenv("DBPORT", endpoint)
	assert.NoError(t, err)

	client, err := CreateClient(os.Getenv("MONGOURI"))
	assert.NotEmpty(t, client)
	assert.NoError(t, err)

	err = client.Ping(context.Background(), nil)
	assert.NoError(t, err)

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

	err = os.Setenv("MONGOURI", fmt.Sprintf("mongodb://%s:%s@%s", os.Getenv("DBUSER"), os.Getenv("DBPASS"), endpoint))
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
