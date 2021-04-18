package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MongoConnection() *mongo.Database {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://admin:secret@localhost:27017"))
	if err != nil {
		panic(err.Error())
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err.Error())
	}
	return client.Database("chat-app", nil)
}

func GetCollection(name string, db *mongo.Database) *mongo.Collection {
	return db.Collection(name, nil)
}
