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

func DeleteOne(ctx context.Context, coll *mongo.Collection, query bson.M) (int64, error) {
	result, err := coll.DeleteOne(ctx, query)
	return result.DeletedCount, err
}

func InsertOne(ctx context.Context, collection *mongo.Collection, model interface{}) (interface{}, bool, error) {
	result, err := collection.InsertOne(ctx, model)
	if err != nil {
		if strings.Index(err.Error(), "duplicate key error collection:") >= 0 {
			return nil, true, nil
		} else {
			return nil, false, err
		}
	} else {
		if idValue, ok := result.InsertedID.(primitive.ObjectID); ok {
			typeOfId := reflect.Indirect(reflect.ValueOf(model)).Type()
			idField := FindIdField(typeOfId)
			mapObjectIdToModel(idValue, model, idField)
		}
		return model, false, err
	}
}

//For Insert
func mapObjectIdToModel(id primitive.ObjectID, model interface{}, idField string) {
	valueOfModel := reflect.Indirect(reflect.ValueOf(model))
	switch valueOfModel.FieldByName(idField).Kind() {
	case reflect.String:
		valueOfModel.FieldByName(idField).Set(reflect.ValueOf(id.String()))
		break
	default:
		valueOfModel.FieldByName(idField).Set(reflect.ValueOf(id))
		break
	}
}

func UpdateOne(ctx context.Context, collection *mongo.Collection, model interface{}, query bson.M) (interface{}, error) { //Patch
	updateQuery := bson.M{
		"$set": model,
	}
	_, err := collection.UpdateOne(ctx, query, updateQuery)
	return model, err
}

func UpsertOne(ctx context.Context, collection *mongo.Collection, model interface{}) (interface{}, error) {
	valueOfModel := reflect.Indirect(reflect.ValueOf(model))
	modelType := valueOfModel.Type()
	idFieldName := FindIdField(modelType)
	idValue := valueOfModel.FieldByName(idFieldName).Interface()
	idType := valueOfModel.FieldByName(idFieldName).Type()
	model1 := reflect.New(modelType).Interface()
	DefaultObjID, _ := primitive.ObjectIDFromHex("000000000000")

	if idValue == "" || idValue == DefaultObjID {
		result, bool, err := InsertOne(ctx, collection, model)
		if bool {
			return nil, err
		} else {
			return result, nil
		}
	} else {
		if idType.String() == "primitive.ObjectID" {
			filter := bson.M{"_id": idValue}
			update := bson.M{
				"$set": model,
			}
			result := collection.FindOneAndUpdate(ctx, filter, update)
			err := result.Decode(&model1)
			return model1, err
		} else {
			isExisted, _ := Exist(ctx, collection, idValue, false)
			if isExisted {
				filter := bson.M{"_id": idValue}
				update := bson.M{
					"$set": model,
				}
				result := collection.FindOneAndUpdate(ctx, filter, update)
				err := result.Decode(&model1)
				return model1, err
			} else {
				result, bool, err := InsertOne(ctx, collection, model)
				if bool {
					return nil, err
				} else {
					return result, nil
				}
			}
		}
	}
}

func PatchOne(ctx context.Context, collection *mongo.Collection, model interface{}, query bson.M) (interface{}, error) {
	updateQuery := bson.M{
		"$set": model,
	}
	_, err := collection.UpdateOne(ctx, query, updateQuery)
	return model, err
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

//For Update
func BuildQueryByIdFromObject(object interface{}, idName string) bson.M {
	var query bson.M
	if v, ok := object.(map[string]interface{}); ok {
		query = bson.M{"_id": v[idName]}
	} else {
		value := reflect.Indirect(reflect.ValueOf(object)).FieldByName(idName).Interface()
		query = bson.M{"_id": value}
	}
	return query
}

//For Patch
func MapToBson(object map[string]interface{}, objectMap map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range object {
		field := objectMap[key]
		result[field] = value
	}
	return result
}

func MakeMapBson(modelType reflect.Type) map[string]string {
	maps := make(map[string]string)
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		key1 := modelType.Field(i).Name
		fields, _ := modelType.FieldByName(key1)
		if tag, ok := fields.Tag.Lookup("bson"); ok {
			if strings.Contains(tag, ",") {
				a := strings.Split(tag, ",")
				maps[key1] = a[0]
			} else {
				maps[key1] = tag
			}
		} else {
			maps[key1] = key1
		}
	}
	return maps
}
