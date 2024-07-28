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
	// ErrMongoNotFound is returned when a given ID cannot be
	// found in the MongoDB collection.
	ErrMongoNotFound = errors.New("document not found")

	// ErrMongoNilColl is returned when a Nil collection is
	// given to NewMongo().
	ErrMongoNilColl = errors.New("collection cannot be nil")

	// ErrMongoFailCreate is returned when MongoDB fails to
	// create a new mailbox entry.
	ErrMongoFailCreate = errors.New("could not submit form, please try again later")

	// ErrMongoFailDelete is returned when MongoDB fails to
	// delete a mailbox entry.
	ErrMongoFailDelete = errors.New("could not delete form, please try again later")

	// ErrMongoInvalidID is returned when referencing an invalid
	// Mongo document ID.
	ErrMongoInvalidID = errors.New("invalid resource id")

	// ErrMongoInternal is returned when an internal database
	// error occurs.
	ErrMongoInternal = errors.New("internal server error")
)

// Mongo implements the Data interface with a MongoDB backend.
type Mongo struct {
	coll *mongodb.Collection
}

// NewMongo initializes a new Mongo Data instance.
func NewMongo(coll *mongodb.Collection) (Data, error) {
	if coll == nil {
		return nil, ErrMongoNilColl
	}
	return &Mongo{coll: coll}, nil
}

// Create a new mailbox entry with the given context and form.
func (m *Mongo) Create(ctx context.Context, f Form) (string, error) {
	res, err := m.coll.InsertOne(ctx, f)
	if err != nil {
		return "", ErrMongoFailCreate
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

// Count the number of elements in the Mongo mailbox collection.
func (m *Mongo) Count(ctx context.Context) int64 {
	count, err := m.coll.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0
	}
	return count
}

// ReadAll returns paginated mailbox entries. It fetches up to 'batch'
// number of elements, after skipping the first (batch * page) elements.
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

// Read the mailbox entry with the given id.
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

// Delete the mailbox entry with the given id.
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
