package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MongoInstance() *mongo.Client {
  if err := godotenv.Load(); err != nil {
    log.Fatal("Error: Unable to load .env file")
  }

  mongoURI := os.Getenv("MOODRING_API_MONGO_URL")
  serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		log.Fatal(err)
    return nil
	}


	fmt.Println("Connected to MongoDB :)")
	return client
}

var Client *mongo.Client = MongoInstance()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("MoodringAPI").Collection(collectionName)
	return collection
}
