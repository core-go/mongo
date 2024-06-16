package mongo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
)

func CreateUniqueIndex(collection *mongo.Collection, fieldName string) (string, error) {
	keys := bson.D{{Key: fieldName, Value: 1}}
	indexName, err := collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    keys,
			Options: options.Index().SetUnique(true),
		},
	)
	return indexName, err
}
func difference(slice1 []string, slice2 []string) []string {
	var diff []string
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			if !found {
				diff = append(diff, s1)
			}
		}
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}
	return diff
}
func FindByIds(ctx context.Context, collection *mongo.Collection, ids []string, result interface{}, idObjectId bool) ([]string, error) {
	var keys []string
	if !idObjectId {
		res := reflect.Indirect(reflect.ValueOf(result))
		find, errFind := collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
		if errFind != nil {
			return ids, errFind
		}
		find.All(ctx, result)
		keySuccess := []string{""}
		//		_, fieldName := FindIdField(modelType)
		for i := 0; i < res.Len(); i++ {
			keySuccess = append(keySuccess, res.Index(i).Field(0).String())
		}
		keys = difference(keySuccess, ids)
		_ = find.Close(ctx)
	} else {
		res := reflect.Indirect(reflect.ValueOf(result))
		id := make([]primitive.ObjectID, 0)
		for _, val := range ids {
			item, err := primitive.ObjectIDFromHex(val)
			if err != nil {
				return nil, err
			}
			id = append(id, item)
		}
		find, err := collection.Find(ctx, bson.M{"_id": bson.M{"$in": id}})
		if err != nil {
			return ids, err
		}
		find.All(ctx, result)
		keySuccess := make([]primitive.ObjectID, 0)
		//_, fieldName := FindIdField(modelType)
		for i := 0; i < res.Len(); i++ {
			key, _ := primitive.ObjectIDFromHex(res.Index(i).Field(0).Interface().(primitive.ObjectID).Hex())
			keySuccess = append(keySuccess, key)
		}
		keyToStr := []string{""}
		for index, _ := range keySuccess {
			key := keySuccess[index].Hex()
			keyToStr = append(keyToStr, key)
		}
		keys = difference(keyToStr, ids)
	}
	if result != nil {
		return keys, nil
	}
	return keys, errors.New("no result return")
}
func PatchMany(ctx context.Context, collection *mongo.Collection, models interface{}, idName string) (*mongo.BulkWriteResult, error) {
	models_ := make([]mongo.WriteModel, 0)
	ids := make([]interface{}, 0)
	switch reflect.TypeOf(models).Kind() {
	case reflect.Slice:
		values := reflect.ValueOf(models)
		length := values.Len()
		if length > 0 {
			if index := findIndex(values.Index(0).Interface(), idName); index != -1 {
				for i := 0; i < length; i++ {
					row := values.Index(i).Interface()
					updateModel := mongo.NewUpdateOneModel().SetUpdate(values.Index(i))
					v, err1 := getValue(row, index)
					if err1 == nil && v != nil {
						if reflect.TypeOf(v).String() != "string" {
							updateModel = mongo.NewUpdateOneModel().SetUpdate(bson.M{
								"$set": row,
							}).SetFilter(bson.M{"_id": v})
						} else {
							if idStr, ok := v.(string); ok {
								updateModel = mongo.NewUpdateOneModel().SetUpdate(bson.M{
									"$set": row,
								}).SetFilter(bson.M{"_id": idStr})
							}
						}
						ids = append(ids, v)
					}
					models_ = append(models_, updateModel)
				}
			}
		}
	}
	var defaultOrdered = false
	return collection.BulkWrite(ctx, models_, &options.BulkWriteOptions{Ordered: &defaultOrdered})
}
func UpdateMaps(ctx context.Context, collection *mongo.Collection, maps []map[string]interface{}, idName string) (*mongo.BulkWriteResult, error) {
	if idName == "" {
		idName = "_id"
	}
	models_ := make([]mongo.WriteModel, 0)
	for _, row := range maps {
		v, _ := row[idName]
		if v != nil {
			updateModel := mongo.NewReplaceOneModel().SetReplacement(bson.M{
				"$set": row,
			}).SetFilter(bson.M{"_id": v})
			models_ = append(models_, updateModel)
		}
	}
	res, err := collection.BulkWrite(ctx, models_)
	return res, err
}
func UpsertMaps(ctx context.Context, collection *mongo.Collection, maps []map[string]interface{}, idName string) (*mongo.BulkWriteResult, error) {
	models_ := make([]mongo.WriteModel, 0)
	for _, row := range maps {
		id, _ := row[idName]
		if id != nil || (reflect.TypeOf(id).String() == "string") || (reflect.TypeOf(id).String() == "string" && len(id.(string)) > 0) {
			updateModel := mongo.NewUpdateOneModel().SetUpdate(bson.M{
				"$set": row,
			}).SetUpsert(true).SetFilter(bson.M{"_id": id})
			models_ = append(models_, updateModel)
		} else {
			insertModel := mongo.NewInsertOneModel().SetDocument(row)
			models_ = append(models_, insertModel)
		}
	}
	res, err := collection.BulkWrite(ctx, models_)
	return res, err
}
