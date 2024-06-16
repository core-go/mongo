package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func FindOne(ctx context.Context, collection *mongo.Collection, query bson.M, result interface{}) (bool, error) {
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
func Find(ctx context.Context, collection *mongo.Collection, query bson.D, arr interface{}) error {
	cur, err := collection.Find(ctx, query)
	if err != nil {
		return err
	}
	er2 := cur.All(ctx, arr)
	return er2
}
