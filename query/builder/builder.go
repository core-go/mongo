package query

import (
	"fmt"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var Operators = map[string]string{
	">=": "$gte",
	">":  "$gt",
	"<=": "$lte",
	"<":  "$lt",
}

func UseQueryByResultType[F any](resultModelType reflect.Type, extract func(F) ([]string, string, []string)) func(filter F) (bson.D, bson.M) {
	b := NewBuilder[F](resultModelType, extract)
	return b.BuildQuery
}
func UseQuery[T any, F any](extract func(F) ([]string, string, []string)) func(filter F) (bson.D, bson.M) {
	var t T
	resultModelType := reflect.TypeOf(t)
	if resultModelType.Kind() == reflect.Ptr {
		resultModelType = resultModelType.Elem()
	}
	b := NewBuilder[F](resultModelType, extract)
	return b.BuildQuery
}

type Builder[F any] struct {
	ModelType reflect.Type
	Extract   func(F) ([]string, string, []string)
}

func NewBuilder[F any](resultModelType reflect.Type, extract func(F) ([]string, string, []string)) *Builder[F] {
	return &Builder[F]{ModelType: resultModelType, Extract: extract}
}
func (b *Builder[F]) BuildQuery(filter F) (bson.D, bson.M) {
	b2 := b.Extract != nil
	if b2 {
		fields, keyword, excluding := b.Extract(filter)
		return Build(filter, b.ModelType, fields, keyword, excluding)
	} else {
		return Build(filter, b.ModelType, nil, "", nil)
	}
}

func Build(filter interface{}, resultModelType reflect.Type, arrFields []string, keyword string, excluding []string) (bson.D, bson.M) {
	var query = bson.D{}
	queryQ := make([]bson.M, 0)
	hasQ := false
	var fields = bson.M{}
	if len(arrFields) > 0 {
		for _, key := range arrFields {
			_, _, qField := getFieldByJson(resultModelType, key)
			if len(qField) <= 0 {
				fields = bson.M{}
				break
			}
			fields[qField] = 1
		}
	}
	value := reflect.Indirect(reflect.ValueOf(filter))
	filterType := value.Type()
	numField := value.NumField()
	for i := 0; i < numField; i++ {
		bsonName := getBson(filterType, i)
		if bsonName == "-" {
			continue
		}
		field := value.Field(i)
		kind := field.Kind()
		x := field.Interface()
		tf := value.Type().Field(i)
		fieldTypeName := tf.Type.String()
		var psv string
		isContinue := false
		if kind == reflect.Ptr {
			if field.IsNil() {
				if fieldTypeName != "*string" {
					continue
				} else {
					isContinue = true
				}
			} else {
				field = field.Elem()
				kind = field.Kind()
				x = field.Interface()
			}
		}
		if !isContinue {
			s0, ok0 := x.(string)
			if ok0 {
				if len(s0) == 0 {
					isContinue = true
				}
				psv = s0
			}
		}

		if len(bsonName) == 0 {
			bsonName = getBsonName(resultModelType, tf.Name)
		}
		if isContinue {
			if len(keyword) > 0 {
				qMatch, isQ := tf.Tag.Lookup("q")
				if isQ {
					hasQ = true
					queryQ1 := bson.M{}
					if qMatch == "=" {
						queryQ1[bsonName] = keyword
					} else if qMatch == "like" {
						queryQ1[bsonName] = primitive.Regex{Pattern: fmt.Sprintf("\\w*%v\\w*", keyword)}
					} else {
						queryQ1[bsonName] = primitive.Regex{Pattern: fmt.Sprintf("^%v", keyword)}
					}
					queryQ = append(queryQ, queryQ1)
				}
			}
			continue
		}
		if len(psv) > 0 {
			key, ok := tf.Tag.Lookup("operator")
			if !ok {
				key, _ = tf.Tag.Lookup("q")
			}
			if key == "=" {
				query = append(query, bson.E{Key: bsonName, Value: psv})
			} else if key == "like" {
				query = append(query, bson.E{Key: bsonName, Value: primitive.Regex{Pattern: fmt.Sprintf("\\w*%v\\w*", psv)}})
			} else {
				query = append(query, bson.E{Key: bsonName, Value: primitive.Regex{Pattern: fmt.Sprintf("^%v", psv)}})
			}
		} else if kind == reflect.Slice {
			if field.Len() > 0 {
				arrQuery := bson.M{}
				arrQuery["$in"] = x
				query = append(query, bson.E{Key: bsonName, Value: arrQuery})
			}
		} else {
			if len(bsonName) > 0 {
				oper, ok1 := tf.Tag.Lookup("operator")
				if ok1 {
					opr, ok2 := Operators[oper]
					if ok2 {
						dQuery := bson.M{}
						dQuery[opr] = x
						query = append(query, bson.E{Key: bsonName, Value: dQuery})
					} else {
						query = append(query, bson.E{Key: bsonName, Value: x})
					}
				} else {
					query = append(query, bson.E{Key: bsonName, Value: x})
				}
			}
		}
	}
	if hasQ {
		query = append(query, bson.E{Key: "$or", Value: queryQ})
	}
	if excluding != nil && len(excluding) > 0 {
		exQuery := bson.M{}
		exQuery["$nin"] = excluding
		query = append(query, bson.E{Key: "_id", Value: exQuery})
	}
	return query, fields
}

func getFieldByJson(modelType reflect.Type, jsonName string) (int, string, string) {
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)
		tag1, ok1 := field.Tag.Lookup("json")
		if ok1 && strings.Split(tag1, ",")[0] == jsonName {
			if tag2, ok2 := field.Tag.Lookup("bson"); ok2 {
				return i, field.Name, strings.Split(tag2, ",")[0]
			}
			return i, field.Name, ""
		}
	}
	return -1, jsonName, jsonName
}
func getBsonName(modelType reflect.Type, fieldName string) string {
	field, found := modelType.FieldByName(fieldName)
	if !found {
		return ""
	}
	if tag, ok := field.Tag.Lookup("bson"); ok {
		return strings.Split(tag, ",")[0]
	}
	return ""
}
func getBson(filterType reflect.Type, i int) string {
	field := filterType.Field(i)
	if tag, ok := field.Tag.Lookup("bson"); ok {
		return strings.Split(tag, ",")[0]
	}
	return ""
}
