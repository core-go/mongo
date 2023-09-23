package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
	"strings"
)

func FindIdField(modelType reflect.Type) (int, string, string) {
	return FindField(modelType, "_id")
}
func FindField(modelType reflect.Type, bsonName string) (int, string, string) {
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)
		bsonTag := field.Tag.Get("bson")
		tags := strings.Split(bsonTag, ",")
		json := field.Name
		if tag1, ok1 := field.Tag.Lookup("json"); ok1 {
			json = strings.Split(tag1, ",")[0]
		}
		for _, tag := range tags {
			if strings.TrimSpace(tag) == bsonName {
				return i, field.Name, json
			}
		}
	}
	return -1, "", ""
}
func Exist(ctx context.Context, collection *mongo.Collection, id interface{}, objectId bool) (bool, error) {
	query := bson.M{"_id": id}
	if objectId {
		objId, err := primitive.ObjectIDFromHex(id.(string))
		if err != nil {
			return false, err
		}
		query = bson.M{"_id": objId}
	}
	x := collection.FindOne(ctx, query)
	if x.Err() != nil {
		if fmt.Sprint(x.Err()) == "mongo: no documents in result" {
			return false, nil
		} else {
			return false, x.Err()
		}
	}
	return true, nil
}

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

func FindOneWithId(ctx context.Context, collection *mongo.Collection, id interface{}, objectId bool, modelType reflect.Type) (interface{}, error) {
	if objectId {
		objId := id.(string)
		return FindOneWithObjectId(ctx, collection, objId, modelType)
	}
	return FindOne(ctx, collection, bson.M{"_id": id}, modelType)
}
func FindOneWithObjectId(ctx context.Context, collection *mongo.Collection, id string, modelType reflect.Type) (interface{}, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return FindOne(ctx, collection, bson.M{"_id": objectId}, modelType)
}
func FindOne(ctx context.Context, collection *mongo.Collection, query bson.M, modelType reflect.Type) (interface{}, error) {
	x := collection.FindOne(ctx, query)
	if x.Err() != nil {
		if fmt.Sprint(x.Err()) == "mongo: no documents in result" {
			return nil, nil
		}
		return nil, x.Err()
	}
	result := reflect.New(modelType).Interface()
	er2 := x.Decode(result)
	if er2 != nil {
		if strings.Contains(fmt.Sprint(er2), "cannot decode") {
			return result, nil
		}
		return nil, er2
	}
	return result, nil
}
