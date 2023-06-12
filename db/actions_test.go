package db

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/mar-cial/items/model"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestInsertItem(t *testing.T) {
	mc, err := CreateClient(os.Getenv("MONGOURI"))
	assert.NoError(t, err)

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
	mc, err := CreateClient(os.Getenv("MONGOURI"))
	assert.NoError(t, err)

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
	mc, err := CreateClient(os.Getenv("MONGOURI"))
	assert.NoError(t, err)

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
	mc, err := CreateClient(os.Getenv("MONGOURI"))
	assert.NoError(t, err)

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
	mc, err := CreateClient(os.Getenv("MONGOURI"))
	assert.NoError(t, err)

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
	mc, err := CreateClient(os.Getenv("MONGOURI"))
	assert.NoError(t, err)

	ctx := context.Background()

	dbname := os.Getenv("DBNAME")
	dbcoll := os.Getenv("DBCOLL")

	coll := mc.Database(dbname).Collection(dbcoll)

	res, err := DeleteOneItem(ctx, coll, ids[0].Hex())
	assert.NoError(t, err)

	fmt.Println(res)
}
