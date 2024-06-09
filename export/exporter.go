package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewExportRepository[T any](db *mongo.Collection,
	buildQuery func(context.Context) bson.D,
	transform func(context.Context, *T) string,
	write func(p []byte) (n int, err error),
	close func() error,
	opts ...func(context.Context) *options.FindOptions,
) *Exporter[T] {
	return NewExporter[T](db, buildQuery, transform, write, close, opts...)
}
func NewExportAdapter[T any](db *mongo.Collection,
	buildQuery func(context.Context) bson.D,
	transform func(context.Context, *T) string,
	write func(p []byte) (n int, err error),
	close func() error,
	opts ...func(context.Context) *options.FindOptions,
) *Exporter[T] {
	return NewExporter[T](db, buildQuery, transform, write, close, opts...)
}
func NewExportService[T any](db *mongo.Collection,
	buildQuery func(context.Context) bson.D,
	transform func(context.Context, *T) string,
	write func(p []byte) (n int, err error),
	close func() error,
	opts ...func(context.Context) *options.FindOptions,
) *Exporter[T] {
	return NewExporter[T](db, buildQuery, transform, write, close, opts...)
}

func NewExporter[T any](db *mongo.Collection,
	buildQuery func(context.Context) bson.D,
	transform func(context.Context, *T) string,
	write func(p []byte) (n int, err error),
	close func() error,
	opts ...func(context.Context) *options.FindOptions,
) *Exporter[T] {
	var opt func(context.Context) *options.FindOptions
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	}
	return &Exporter[T]{Collection: db, Write: write, Close: close, Transform: transform, BuildQuery: buildQuery, BuildFindOptions: opt}
}

type Exporter[T any] struct {
	Collection       *mongo.Collection
	BuildQuery       func(context.Context) bson.D
	BuildFindOptions func(context.Context) *options.FindOptions
	Transform        func(context.Context, *T) string
	Write            func(p []byte) (n int, err error)
	Close            func() error
}

func (s *Exporter[T]) Export(ctx context.Context) (int64, error) {
	query := s.BuildQuery(ctx)
	var cursor *mongo.Cursor
	var err error
	if s.BuildFindOptions != nil {
		optionsFind := s.BuildFindOptions(ctx)
		cursor, err = s.Collection.Find(ctx, query, optionsFind)
	} else {
		cursor, err = s.Collection.Find(ctx, query)
	}
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)
	var i int64
	i = 0
	for cursor.Next(ctx) {
		var obj T
		err = cursor.Decode(&obj)
		if err != nil {
			return i, err
		}
		err1 := s.TransformAndWrite(ctx, s.Write, &obj)
		if err1 != nil {
			return i, err1
		}
		i = i + 1
	}
	return i, cursor.Err()
}

func (s *Exporter[T]) TransformAndWrite(ctx context.Context, write func(p []byte) (n int, err error), model *T) error {
	line := s.Transform(ctx, model)
	_, er := write([]byte(line))
	return er
}
