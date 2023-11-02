package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

func NewClient(ctx context.Context, database string) (db *mongo.Database, bucket *gridfs.Bucket, err error) {
	mongoDBURL := os.Getenv("MONGO_URL")
	//serverAPI := options.ServerAPI(options.ser)
	//clientOptions := options.Client().ApplyURI(mongoDBURL).SetServerAPIOptions(serverAPI)
	//
	//client, err := mongo.Connect(context.TODO(), clientOptions)
	clientOptions := options.Client().ApplyURI(mongoDBURL)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to mongoDB: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	bucket, err = gridfs.NewBucket(client.Database(database))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create bucket in GRIDFS: %v", err)
	}
	log.Println("Successfully connected to database")
	return client.Database(database), bucket, nil
}
