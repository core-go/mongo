package mongo

import "context"

type SearchBuilder interface {
	Search(ctx context.Context, searchModel interface{}) (interface{}, int64, error)
}
