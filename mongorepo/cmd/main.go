package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bastianrob/go-experiences/mongorepo/pkg/models"
	"github.com/bastianrob/go-experiences/mongorepo/pkg/mongorepo"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// inisialisasi koneksi ke mongo
	ctx := context.Background()
	conn := os.Getenv("MONGO_CONN")
	mongoop := options.Client().ApplyURI(conn)
	mongocl, _ := mongo.Connect(ctx, mongoop)
	mongodb := mongocl.Database("dbtest")

	// personRepo, ngarah ke collection 'person'
	personRepo := mongorepo.New(
		mongodb.Collection("person"),
		func() interface{} {
			return &models.Person{}
		})

	// enemyRepo, ngarah ke collection 'enemy'
	enemyRepo := mongorepo.New(
		mongodb.Collection("enemy"),
		func() interface{} {
			return &models.Enemy{}
		})

	fmt.Println(personRepo, enemyRepo)
}
