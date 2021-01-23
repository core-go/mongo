package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type MongoBatchPatcher struct {
	collection *mongo.Collection
	IdName     string
	modelType  reflect.Type
	modelsType reflect.Type
}

func NewMongoBatchPatcherWithIdName(database *mongo.Database, collectionName string, modelType reflect.Type, fieldName string) *MongoBatchPatcher {
	if len(fieldName) == 0 {
		_, idName := FindIdField(modelType)
		fieldName = idName
	}
	return CreateMongoBatchPatcherIdName(database, collectionName, modelType, fieldName)
}

func NewMongoBatchPatcher(database *mongo.Database, collectionName string, modelType reflect.Type) *MongoBatchPatcher {
	return CreateMongoBatchPatcherIdName(database, collectionName, modelType, "")
}

func CreateMongoBatchPatcherIdName(database *mongo.Database, collectionName string, modelType reflect.Type, fieldName string) *MongoBatchPatcher {
	modelsType := reflect.Zero(reflect.SliceOf(modelType)).Type()
	collection := database.Collection(collectionName)
	return &MongoBatchPatcher{collection, fieldName, modelType, modelsType}
}

func (w *MongoBatchPatcher) WriteBatch(ctx context.Context, models []map[string]interface{}) ([]int, []int, error) {
	successIndices := make([]int, 0)
	failIndices := make([]int, 0)

	s := reflect.ValueOf(models)
	_, err := PatchMaps(ctx, w.collection, models, w.IdName)

	if err == nil {
		// Return full success
		for i := 0; i < s.Len(); i++ {
			successIndices = append(successIndices, i)
		}
		return successIndices, failIndices, err
	}

	if bulkWriteException, ok := err.(mongo.BulkWriteException); ok {
		for _, writeError := range bulkWriteException.WriteErrors {
			failIndices = append(failIndices, writeError.Index)
		}
		for i := 0; i < s.Len(); i++ {
			if !InArray(i, failIndices) {
				successIndices = append(successIndices, i)
			}
		}
	} else {
		// Return full fail
		for i := 0; i < s.Len(); i++ {
			failIndices = append(failIndices, i)
		}
	}
	return successIndices, failIndices, err
}
