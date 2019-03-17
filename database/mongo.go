package database

import (
	"context"
	"os"

	"github.com/gomodule/redigo/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongo(ctx context.Context) (*mongo.Client, error) {
	uri := "mongodb://localhost:27017"

	if val, set := os.LookupEnv("CASHBOT_MONGO_URI"); set {
		uri = val
	}

	client, err := mongo.NewClient(
		options.Client().ApplyURI(uri),
	)
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func ConnectRedis() (redis.Conn, error) {
	uri := "redis://localhost:6379"

	if val, set := os.LookupEnv("CASHBOT_REDIS_URI"); set {
		uri = val
	}

	conn, err := redis.DialURL(uri)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
