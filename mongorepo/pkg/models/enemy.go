package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Enemy is a database object
type Enemy struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name"`
}
