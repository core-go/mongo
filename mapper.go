package mongo

import "context"

type Mapper interface {
	//BsonToLocation(ctx context.Context, model interface{}) (interface{}, error)
	DbToModel(ctx context.Context, model interface{}) (interface{}, error)
	DbToModels(ctx context.Context, model interface{}) (interface{}, error)

	ModelToDb(ctx context.Context, model interface{}) (interface{}, error)
	ModelsToDb(ctx context.Context, model interface{}) (interface{}, error)
}
