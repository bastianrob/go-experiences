package mongorepo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	virtualDelete = bson.M{"$set": bson.M{"deleted": true}}
)

// MongoRepo base class
type MongoRepo struct {
	collection  *mongo.Collection
	constructor func() interface{}
}

// New creates a new instance of MongoRepo
func New(coll *mongo.Collection, cons func() interface{}) *MongoRepo {
	return &MongoRepo{
		collection:  coll,
		constructor: cons,
	}
}

// Get a list of resource
func (r *MongoRepo) Get(ctx context.Context) ([]interface{}, error) {
	cur, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var result []interface{}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		entry := r.constructor()
		if err = cur.Decode(entry); err != nil {
			return nil, err
		}
		result = append(result, entry)
	}

	return result, nil
}

// GetOne resource based on its ID
func (r *MongoRepo) GetOne(ctx context.Context, id string) (interface{}, error) {
	_id, _ := primitive.ObjectIDFromHex(id)
	res := r.collection.FindOne(ctx, bson.M{"_id": _id})
	dbo := r.constructor()
	err := res.Decode(dbo)
	return dbo, err
}

// Create a new resource
func (r *MongoRepo) Create(ctx context.Context, obj interface{}) error {
	_, err := r.collection.InsertOne(ctx, obj)
	if err != nil {
		return err
	}

	return nil
}

// Update a resource
func (r *MongoRepo) Update(ctx context.Context, id string, obj interface{}) error {
	_id, _ := primitive.ObjectIDFromHex(id)
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": _id}, obj)
	if err != nil {
		return err
	}

	return nil
}

// Delete a resource, virtually by marking it as {"deleted": true}
func (r *MongoRepo) Delete(ctx context.Context, id string) error {
	_id, _ := primitive.ObjectIDFromHex(id)
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": _id}, virtualDelete)
	if err != nil {
		return err
	}

	return nil
}
