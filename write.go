package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"strings"
)

func MakeBsonMap(modelType reflect.Type) map[string]string {
	maps := make(map[string]string)
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)
		key1 := field.Name
		if tag0, ok0 := field.Tag.Lookup("json"); ok0 {
			if strings.Contains(tag0, ",") {
				a := strings.Split(tag0, ",")
				key1 = a[0]
			} else {
				key1 = tag0
			}
		}
		if tag, ok := field.Tag.Lookup("bson"); ok {
			if tag != "-" {
				if strings.Contains(tag, ",") {
					a := strings.Split(tag, ",")
					if key1 == "-" {
						key1 = a[0]
					}
					maps[key1] = a[0]
				} else {
					if key1 == "-" {
						key1 = tag
					}
					maps[key1] = tag
				}
			}
		} else {
			if key1 == "-" {
				key1 = field.Name
			}
			maps[key1] = key1
		}
	}
	return maps
}

func InsertOne(ctx context.Context, collection *mongo.Collection, model interface{}) (*primitive.ObjectID, int64, error) {
	result, err := collection.InsertOne(ctx, model)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "duplicate key error collection:") {
			return nil, 0, nil
		} else {
			return nil, 0, err
		}
	} else {
		if idValue, ok := result.InsertedID.(primitive.ObjectID); ok {
			return &idValue, 1, nil
		}
		return nil, 1, nil
	}
}

// For Patch
func MapToBson(object map[string]interface{}, objectMap map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range object {
		field, ok := objectMap[key]
		if ok {
			result[field] = value
		}
	}
	return result
}
func PatchOne(ctx context.Context, collection *mongo.Collection, id interface{}, model map[string]interface{}) (int64, error) {
	filter := bson.M{"_id": id}
	updateQuery := bson.M{
		"$set": model,
	}
	result, err := collection.UpdateOne(ctx, filter, updateQuery)
	if err != nil {
		return 0, err
	}
	if result.ModifiedCount > 0 {
		return result.ModifiedCount, err
	} else if result.UpsertedCount > 0 {
		return result.UpsertedCount, err
	} else {
		return result.MatchedCount, err
	}
}
func PatchOneByFilter(ctx context.Context, collection *mongo.Collection, filter bson.D, model map[string]interface{}) (int64, error) { //Patch
	updateQuery := bson.M{
		"$set": model,
	}
	result, err := collection.UpdateOne(ctx, filter, updateQuery)
	if result.ModifiedCount > 0 {
		return result.ModifiedCount, err
	} else if result.UpsertedCount > 0 {
		return result.UpsertedCount, err
	} else {
		return result.MatchedCount, err
	}
}
func UpdateOne(ctx context.Context, collection *mongo.Collection, id interface{}, model interface{}) (int64, error) { //Patch
	filter := bson.M{"_id": id}
	updateQuery := bson.M{
		"$set": model,
	}
	result, err := collection.UpdateOne(ctx, filter, updateQuery)
	if result.ModifiedCount > 0 {
		return result.ModifiedCount, err
	} else if result.UpsertedCount > 0 {
		return result.UpsertedCount, err
	} else {
		return result.MatchedCount, err
	}
}
func UpdateOneByFilter(ctx context.Context, collection *mongo.Collection, filter bson.D, model interface{}) (int64, error) { //Patch
	updateQuery := bson.M{
		"$set": model,
	}
	result, err := collection.UpdateOne(ctx, filter, updateQuery)
	if result.ModifiedCount > 0 {
		return result.ModifiedCount, err
	} else if result.UpsertedCount > 0 {
		return result.UpsertedCount, err
	} else {
		return result.MatchedCount, err
	}
}
func UpsertOne(ctx context.Context, collection *mongo.Collection, id interface{}, model interface{}) (int64, error) {
	filter := bson.M{"_id": id}
	updateQuery := bson.M{
		"$set": model,
	}
	opts := options.Update().SetUpsert(true)
	res, err := collection.UpdateOne(ctx, filter, updateQuery, opts)
	if res.ModifiedCount > 0 {
		return res.ModifiedCount, err
	} else if res.UpsertedCount > 0 {
		return res.UpsertedCount, err
	} else {
		return res.MatchedCount, err
	}
}
func UpsertOneByFilter(ctx context.Context, collection *mongo.Collection, filter bson.D, model interface{}) (int64, error) {
	updateQuery := bson.M{
		"$set": model,
	}
	opts := options.Update().SetUpsert(true)
	res, err := collection.UpdateOne(ctx, filter, updateQuery, opts)
	if res.ModifiedCount > 0 {
		return res.ModifiedCount, err
	} else if res.UpsertedCount > 0 {
		return res.UpsertedCount, err
	} else {
		return res.MatchedCount, err
	}
}
func DeleteOne(ctx context.Context, collection *mongo.Collection, id interface{}) (int64, error) {
	filter := bson.M{"_id": id}
	result, err := collection.DeleteOne(ctx, filter)
	if result == nil {
		return 0, err
	}
	return result.DeletedCount, err
}
