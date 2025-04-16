package database

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var MongoDB *mongo.Database

func InitMongoDB() {
	_ = godotenv.Load() // Load .env file

	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB_NAME")

	if uri == "" || dbName == "" {
		log.Fatal("Missing MONGO_URI or MONGO_DB_NAME in environment variables")
	}

	clientOpts := options.Client().ApplyURI(uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
	}

	// Ping the DB to confirm connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB ping failed: %v", err)
	}

	MongoClient = client
	MongoDB = client.Database(dbName)

	index := mongo.IndexModel{
		Keys: bson.D{{Key: "location", Value: "2dsphere"}},
	}
	_, _ = MongoDB.Collection("users").Indexes().CreateOne(ctx, index)

	log.Println("✅ MongoDB connected successfully")
}

// Uniqueness/index enforcement to likes and matches
func EnsureIndexes(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	likesIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "fromUser", Value: 1}, {Key: "toUser", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	matchesIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "user1", Value: 1}, {Key: "user2", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := db.Collection("likes").Indexes().CreateOne(ctx, likesIndex)
	if err != nil {
		return err
	}

	_, err = db.Collection("matches").Indexes().CreateOne(ctx, matchesIndex)
	if err != nil {
		return err
	}

	return nil
}
