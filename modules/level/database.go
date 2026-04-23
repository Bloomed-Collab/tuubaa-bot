package level

import (
	"context"
	"time"

	"github.com/S42yt/tuubaa-bot/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const lvlCollection = "levels"

type levelEntry struct {
	UserID string `bson:"user_id"`
	XP     int64  `bson:"xp"`
}

func getXP(userID string) (int64, error) {
	db := core.NewMongoHandler()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Connect(ctx); err != nil {
		return 0, err
	}
	defer db.Disconnect(context.Background())

	var entry levelEntry
	err := db.Collection(lvlCollection).FindOne(ctx, bson.M{"user_id": userID}).Decode(&entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, nil
		}
		return 0, err
	}

	return entry.XP, nil
}

func getAllXP() ([]levelEntry, error) {
	db := core.NewMongoHandler()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Connect(ctx); err != nil {
		return nil, err
	}
	defer db.Disconnect(context.Background())

	opts := options.Find().SetSort(bson.D{{Key: "xp", Value: -1}})
	cursor, err := db.Collection(lvlCollection).Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entries []levelEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

func upsertXP(userID string, xp int64) error {
	db := core.NewMongoHandler()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Connect(ctx); err != nil {
		return err
	}
	defer db.Disconnect(context.Background())

	opts := options.UpdateOne().SetUpsert(true)
	_, err := db.Collection(lvlCollection).UpdateOne(ctx,
		bson.M{"user_id": userID},
		bson.M{"$set": bson.M{"xp": xp}},
		opts,
	)

	if err != nil {
		return err
	}

	return nil
}
