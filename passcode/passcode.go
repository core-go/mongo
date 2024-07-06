package passcode

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PasscodeRepository struct {
	collection    *mongo.Collection
	passcodeName  string
	expiredAtName string
}

func NewPasscodeRepository(db *mongo.Database, collectionName string, options ...string) *PasscodeRepository {
	var passcodeName, expiredAtName string
	if len(options) >= 1 && len(options[0]) > 0 {
		expiredAtName = options[0]
	} else {
		expiredAtName = "expiredAt"
	}
	if len(options) >= 2 && len(options[1]) > 0 {
		passcodeName = options[1]
	} else {
		passcodeName = "passcode"
	}
	return &PasscodeRepository{db.Collection(collectionName), passcodeName, expiredAtName}
}

func (p *PasscodeRepository) Save(ctx context.Context, id string, passcode string, expiredAt time.Time) (int64, error) {
	pass := make(map[string]interface{})
	pass["_id"] = id
	pass[p.passcodeName] = passcode
	pass[p.expiredAtName] = expiredAt
	updateQuery := bson.M{
		"$set": pass,
	}
	filter := bson.M{"_id": id}
	opts := options.Update().SetUpsert(true)
	res, err := p.collection.UpdateOne(ctx, filter, updateQuery, opts)
	if res.ModifiedCount > 0 {
		return res.ModifiedCount, err
	} else if res.UpsertedCount > 0 {
		return res.UpsertedCount, err
	} else {
		return res.MatchedCount, err
	}
}

func (p *PasscodeRepository) Load(ctx context.Context, id string) (string, time.Time, error) {
	idQuery := bson.M{"_id": id}
	x := p.collection.FindOne(ctx, idQuery)
	er1 := x.Err()
	if er1 != nil {
		if strings.Compare(fmt.Sprint(er1), "mongo: no documents in result") == 0 {
			return "", time.Now().Add(-24 * time.Hour), nil
		}
		return "", time.Now().Add(-24 * time.Hour), er1
	}
	k, er3 := x.DecodeBytes()
	if er3 != nil {
		return "", time.Now().Add(-24 * time.Hour), er3
	}

	code := strings.Trim(k.Lookup(p.passcodeName).String(), "\"")
	expiredAt := k.Lookup(p.expiredAtName).Time()
	return code, expiredAt, nil
}

func (p *PasscodeRepository) Delete(ctx context.Context, id string) (int64, error) {
	filter := bson.M{"_id": id}
	result, err := p.collection.DeleteOne(ctx, filter)
	if result == nil {
		return 0, err
	}
	return result.DeletedCount, err
}
