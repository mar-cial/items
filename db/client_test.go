package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

var endpoint string

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
		err = mongoC.Terminate(ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}()
}
