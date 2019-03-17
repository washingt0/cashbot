package database

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Entry struct {
	Value       float64   `json:"value"`
	Payment     bool      `json:"payment"`
	Owner       string    `json:"user"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`
}

func (e *Entry) AddEntry(mgo *mongo.Client) error {
	coll := mgo.Database("cashbot").Collection("entries")
	_, err := coll.InsertOne(context.Background(), e)
	if err != nil {
		return err
	}
	return nil
}

func (e *Entry) GetOwnerEntries(mgo *mongo.Client) ([]Entry, error) {
	if e.Owner == "" {
		return []Entry{}, errors.New("Sorry, I didn't recognize you")
	}

	coll := mgo.Database("cashbot").Collection("entries")
	cur, err := coll.Find(context.Background(), bson.D{
		{"owner", e.Owner},
	}, &options.FindOptions{
		Sort: bson.D{
			{"createdat", 1},
		},
	})
	if err != nil {
		return []Entry{}, err
	}
	defer cur.Close(context.Background())

	var data []Entry
	for cur.Next(context.Background()) {
		var elem Entry
		if err := cur.Decode(&elem); err != nil {
			return []Entry{}, err
		}

		data = append(data, elem)
	}

	if err := cur.Err(); err != nil {
		return []Entry{}, err
	}

	return data, nil
}

func (e *Entry) DropEntries(mgo *mongo.Client) error {
	if e.Owner == "" {
		return errors.New("Sorry, I didn't recognize you")
	}
	coll := mgo.Database("cashbot").Collection("entries")

	_, err := coll.DeleteMany(context.Background(), bson.D{
		{"owner", e.Owner},
	})
	if err != nil {
		return err
	}
	return nil
}
