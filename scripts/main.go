package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type jsonEntry struct {
	UserID string `json:"userId"`
	XP     int64  `json:"dezixp"`
}

func main() {
	godotenv.Load() //nolint:errcheck

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	mongoDB := os.Getenv("MONGO_DB")
	if mongoDB == "" {
		mongoDB = "tuubaa"
	}

	filePath := "level.json"
	if len(os.Args) > 1 {
		filePath = os.Args[1]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(context.Background()) //nolint:errcheck

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("ping mongodb at %s: %v", mongoURI, err)
	}
	fmt.Printf("Connected to %s / %s\n\n", mongoURI, mongoDB)

	coll := client.Database(mongoDB).Collection("levels")

	f, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("open %s: %v", filePath, err)
	}
	defer f.Close()

	var imported, skipped int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry jsonEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			log.Printf("  skip (bad json): %v", err)
			skipped++
			continue
		}
		if entry.UserID == "" {
			skipped++
			continue
		}

		uCtx, uCancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := coll.UpdateOne(uCtx,
			bson.M{"user_id": entry.UserID},
			bson.M{"$set": bson.M{"user_id": entry.UserID, "xp": entry.XP}},
			options.UpdateOne().SetUpsert(true),
		)
		uCancel()

		if err != nil {
			log.Printf("  skip %s (db error): %v", entry.UserID, err)
			skipped++
			continue
		}

		fmt.Printf("  ✓  %-22s %d xp\n", entry.UserID, entry.XP)
		imported++
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("scan: %v", err)
	}

	fmt.Printf("\nDone — %d imported, %d skipped\n", imported, skipped)
}
