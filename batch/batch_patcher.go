package batch

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type BatchPatcher struct {
	collection *mongo.Collection
	IdName     string
}

func NewBatchPatcherWithId(database *mongo.Database, collectionName string, fieldName string) *BatchPatcher {
	return CreateMongoBatchPatcherIdName(database, collectionName, fieldName)
}

func NewBatchPatcher(database *mongo.Database, collectionName string) *BatchPatcher {
	return CreateMongoBatchPatcherIdName(database, collectionName, "")
}

func CreateMongoBatchPatcherIdName(database *mongo.Database, collectionName string, fieldName string) *BatchPatcher {
	collection := database.Collection(collectionName)
	return &BatchPatcher{collection, fieldName}
}

func (w *BatchPatcher) Write(ctx context.Context, models []map[string]interface{}) ([]int, error) {
	failIndices := make([]int, 0)
	_, err := PatchMaps(ctx, w.collection, models, w.IdName)

	if err == nil {
		return failIndices, err
	}

	if bulkWriteException, ok := err.(mongo.BulkWriteException); ok {
		for _, writeError := range bulkWriteException.WriteErrors {
			failIndices = append(failIndices, writeError.Index)
		}
	}
	return failIndices, err
}
