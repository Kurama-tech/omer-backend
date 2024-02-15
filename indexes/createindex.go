package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	Name   string `bson:"name"`
	UserID string `bson:"userid"`
}

func main() {
	// Set up MongoDB client
	clientOptions := options.Client().ApplyURI("<MongoURL>")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// Access users collection
	usersCollection := client.Database("omer").Collection("customer")
	

	// Create unique index on userid field
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"number": 1},
		Options: options.Index().SetUnique(true),
	}
	indexName, err := usersCollection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created index:", indexName)

	
}
