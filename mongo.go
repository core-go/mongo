package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"strings"
)

func SetupMongo(ctx context.Context, mongoConfig MongoConfig) (*mongo.Database, error) {
	return CreateConnection(ctx, mongoConfig.Uri, mongoConfig.Database)
}

func CreateConnection(ctx context.Context, uri string, database string) (*mongo.Database, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	db := client.Database(database)
	return db, nil
}

func FindOneWithId(ctx context.Context, collection *mongo.Collection, id interface{}, objectId bool, modelType reflect.Type) (interface{}, error) {
	if objectId {
		objId := id.(string)
		return FindOneWithObjectId(ctx, collection, objId, modelType)
	} else {
		return FindOne(ctx, collection, bson.M{"_id": id}, modelType)
	}
}

func FindOneWithObjectId(ctx context.Context, collection *mongo.Collection, id string, modelType reflect.Type) (interface{}, error) {
	objectId, _ := primitive.ObjectIDFromHex(id)
	return FindOne(ctx, collection, bson.M{"_id": objectId}, modelType)
}

func FindOne(ctx context.Context, collection *mongo.Collection, query bson.M, modelType reflect.Type) (interface{}, error) {
	x := collection.FindOne(ctx, query)
	if x.Err() != nil {
		if strings.Compare(fmt.Sprint(x.Err()), "mongo: no documents in result") == 0 {
			return nil, nil
		} else {
			return nil, x.Err()
		}
	} else {
		result := reflect.New(modelType).Interface()
		er2 := x.Decode(result)
		if er2 != nil {
			if strings.Contains(fmt.Sprint(er2), "cannot decode null into") {
				return result, nil
			}
			return nil, er2
		} else {
			return result, nil
		}
	}
}

func Find(ctx context.Context, collection *mongo.Collection, query bson.M, modelType reflect.Type) (interface{}, error) {
	cur, err := collection.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	modelsType := reflect.Zero(reflect.SliceOf(modelType)).Type()
	arr := reflect.New(modelsType).Interface()
	er2 := cur.All(ctx, arr)
	return arr, er2
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
		if strings.Compare(fmt.Sprint(x.Err()), "mongo: no documents in result") == 0 {
			return false, nil
		} else {
			return false, x.Err()
		}
	}
	return true, nil
}

//For Get By Id
func FindIdField(modelType reflect.Type) string {
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)
		bsonTag := field.Tag.Get("bson")
		tags := strings.Split(bsonTag, ",")
		for _, tag := range tags {
			if strings.Compare(strings.TrimSpace(tag), "_id") == 0 {
				return field.Name
			}
		}
	}
	return ""
}

//For Search and Patch
func GetBsonColumnName(ModelType reflect.Type, fieldName string) string {
	fields, _ := ModelType.FieldByName(fieldName)
	if tag, ok := fields.Tag.Lookup("bson"); ok {
		return tag
	} else {
		return fieldName
	}
}
