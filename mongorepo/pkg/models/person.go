package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Person is a database object
type Person struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name"`
}
