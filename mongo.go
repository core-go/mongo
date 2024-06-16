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
