package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func FindOneAndDecode(ctx context.Context, collection *mongo.Collection, query bson.M, result interface{}) (bool, error) {
	x := collection.FindOne(ctx, query)
	if x.Err() != nil {
		if fmt.Sprint(x.Err()) == "mongo: no documents in result" {
			return false, nil
		}
		return false, x.Err()
	}
	er2 := x.Decode(result)
	return true, er2
}
func FindAndDecode(ctx context.Context, collection *mongo.Collection, query bson.M, arr interface{}) (bool, error) {
	cur, err := collection.Find(ctx, query)
	if err != nil {
		return false, err
	}
	er2 := cur.All(ctx, arr)
	return true, er2
}
