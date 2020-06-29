package mongo

import (
	"context"
	"github.com/common-go/search"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type SearchResultBuilder interface {
	BuildSearchResult(ctx context.Context, collection *mongo.Collection, searchModel interface{}, modelType reflect.Type) (*search.SearchResult, error)
}
