package config

import (
	"context" // Required for context
	"log"
	"time"    // Required for time

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options" // Required for options
)

// Declare the Client variable globally but don't initialize it here
var Client *mongo.Client

// ConnectDB initializes the MongoDB connection
func ConnectDB() *mongo.Client {
	log.Println("Connecting to MongoDB...")

    // IMPORTANT: Add your connection URI here
	ClientOptions := options.Client().ApplyURI("mongodb+srv://Morayo:O3gaf7l9aNFeXRLn@cluster0.vtsxqud.mongodb.net/?appName=Cluster0")
	
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	defer cancel()

	var err error // Declare err here for scope

	Client, err = mongo.Connect(ctx, ClientOptions)

	if err != nil {
		log.Fatalf("Failed to connect to MongoDB %v", err)
	}

    // Fixed typo: Pimg -> Ping
	err = Client.Ping(ctx, nil) 
	if err != nil {
		log.Fatalf("MongoDB is unreachable %v", err)
	}

    // Simplified connection check and return
	log.Println("Successfully connected to MongoDB!")
	return Client
}

// OpenCollection returns a handle to a specific collection
func OpenCollection(collectionName string) *mongo.Collection {

	if Client == nil { // Fixed variable name case: client -> Client
		log.Fatal("MongoDB is not initialized. Please connect db first")
	}

    // Fixed type: *mongo.connection -> *mongo.Collection (return type)
	return Client.Database("userdb").Collection(collectionName) 
}
