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

func (e *Entry) GetOwnerEntries(mgo *mongo.Client, start *time.Time, end *time.Time) ([]Entry, error) {
	if e.Owner == "" {
		return []Entry{}, errors.New("Sorry, I didn't recognize you")
	}

	var filter bson.D

	if start == nil && end == nil {
		filter = bson.D{
			{"owner", e.Owner},
		}
	} else {
		filter = bson.D{
			{"owner", e.Owner},
			{"createdat",
				bson.D{
					{"$gte", start},
					{"$lt", end},
				},
			},
		}
	}

	coll := mgo.Database("cashbot").Collection("entries")
	cur, err := coll.Find(context.Background(), filter, &options.FindOptions{
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

func (e *Entry) DropLastEntry(mgo *mongo.Client) error {
	if e.Owner == "" {
		return errors.New("Sorry, I didn't recognize you")
	}
	coll := mgo.Database("cashbot").Collection("entries")

	doc := coll.FindOne(context.Background(), bson.D{
		{"owner", e.Owner},
	}, &options.FindOneOptions{Sort: bson.D{
		{"createdat", -1},
	}})

	if bytes, err := doc.DecodeBytes(); err != nil {
		return err
	} else {
		_id := bytes.Lookup("_id").ObjectID()

		if _, err := coll.DeleteOne(context.Background(), bson.D{
			{"_id", _id},
		}); err != nil {
			return err
		}
	}
	return nil
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
