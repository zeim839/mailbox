package data

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrMongoNotFound   = errors.New("document not found")
	ErrMongoNilColl    = errors.New("collection cannot be nil")
	ErrMongoFailCreate = errors.New("could not submit form, please try again later")
	ErrMongoFailUpdate = errors.New("could not update form, please try again later")
	ErrMongoFailDelete = errors.New("could not delete form, please try again later")
	ErrMongoInvalidID  = errors.New("invalid resource id")
	ErrMongoInternal   = errors.New("internal server error")
)

type Mongo struct {
	coll *mongodb.Collection
}

func NewMongo(coll *mongodb.Collection) (*Mongo, error) {
	if coll == nil {
		return nil, ErrMongoNilColl
	}
	return &Mongo{coll: coll}, nil
}

func (m *Mongo) Create(ctx context.Context, f Form) (string, error) {
	res, err := m.coll.InsertOne(ctx, f)
	if err != nil {
		return "", ErrMongoFailCreate
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (m *Mongo) Count(ctx context.Context) int64 {
	count, err := m.coll.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0
	}
	return count
}

func (m *Mongo) ReadAll(ctx context.Context, batch, page int64) ([]Form, error) {
	cursor, err := m.coll.Find(ctx, bson.D{},
		options.Find().SetLimit(batch).SetSkip(page*batch))

	if err != nil {
		return []Form{}, ErrMongoNotFound
	}

	result := []Form{}
	if err := cursor.All(ctx, &result); err != nil {
		return []Form{}, ErrMongoInternal
	}

	return result, nil
}

func (m *Mongo) Read(ctx context.Context, id string) (Form, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return Form{}, ErrMongoInvalidID
	}

	var form Form
	err = m.coll.FindOne(ctx,
		bson.D{{Key: "_id", Value: objID}}).
		Decode(&form)

	if err != nil {
		if err == mongodb.ErrNoDocuments {
			return Form{}, ErrMongoNotFound
		}
		return Form{}, ErrMongoInternal
	}

	return form, nil
}

func (m *Mongo) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrMongoInvalidID
	}

	res, err := m.coll.DeleteOne(ctx,
		bson.D{{Key: "_id", Value: objID}})

	if err != nil {
		return ErrMongoFailDelete
	}

	if res.DeletedCount == 0 {
		return ErrMongoNotFound
	}

	return nil
}
