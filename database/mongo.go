package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Mongo contains the mongo client and database instances
type Mongo struct {
	client *mongo.Client
	db     *mongo.Database
}

// MongoImpl defines the interface for interacting with MongoDB
type MongoImpl interface {
	Cursor(collection, input string) (*mongo.Cursor, error)
	FindOne(collection string, filter interface{}) *mongo.SingleResult
	Upsert(collection string, filter interface{}, update interface{}) *mongo.SingleResult
	FindOneAndDelete(collection string, filter interface{}) *mongo.SingleResult
	Disconnect() error
}

func (m *Mongo) Cursor(collection, input string) (*mongo.Cursor, error) {
	return m.db.Collection(collection).Find(context.Background(), bson.M{"name": bson.M{"$regex": fmt.Sprintf("^%s", input), "$options": "i"}}, options.Find().SetProjection(bson.M{"name": 1}))
}

// FindOne finds a single document that matches the filter in the given collection
func (m *Mongo) FindOne(collection string, filter interface{}) *mongo.SingleResult {
	return m.db.Collection(collection).FindOne(context.Background(), filter)
}

// Upsert updates a document if it already exists or inserts it otherwise
func (m *Mongo) Upsert(collection string, filter interface{}, update interface{}) *mongo.SingleResult {
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	return m.db.Collection(collection).FindOneAndUpdate(context.Background(), filter, update, opts)
}

// FindOneAndDelete deletes a document that matches the filter in the given collection
func (m *Mongo) FindOneAndDelete(collection string, filter interface{}) *mongo.SingleResult {
	return m.db.Collection(collection).FindOneAndDelete(context.Background(), filter)
}

// Disconnect disconnects the mongo client
func (m *Mongo) Disconnect() error {
	return m.client.Disconnect(context.Background())
}

// Setup connects to the MongoDB database and returns a Mongo instance
func Setup(database string) (MongoImpl, error) {
	var mdb MongoImpl
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 160*time.Second)
		defer cancel()
		client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %s", err)
		}

		mdb = &Mongo{
			client: client,
			db:     client.Database(database),
		}
	}()
	// Wait for MongoDB connection to be established
	for mdb == nil {
		time.Sleep(100 * time.Millisecond)
	}
	return mdb, nil
}
