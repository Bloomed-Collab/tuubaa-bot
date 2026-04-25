package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const oldFactor = 200.0

const (
	newLevelStepFactor = int64(50)
	newLvlMax          = 1000
)

type jsonEntry struct {
	UserID string `json:"userId"`
	OldXP  int64  `json:"dezixp"`
}

func oldCalcLevel(xp int64) int {
	if xp <= 0 {
		return 0
	}
	level := int(math.Floor(math.Sqrt(float64(xp) / oldFactor)))
	if level > newLvlMax {
		return newLvlMax
	}
	return level
}

func newTotalXPForLevel(level int) int64 {
	if level <= 0 {
		return 0
	}
	l := int64(level)
	return (newLevelStepFactor / 2) * l * (l + 1)
}

func main() {
	godotenv.Load()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	mongoDB := os.Getenv("MONGO_DB")
	if mongoDB == "" {
		mongoDB = "tuubaa"
	}

	filePath := "levels.json"
	if len(os.Args) > 1 {
		filePath = os.Args[1]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(context.Background())

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
			skipped++
			continue
		}
		if entry.UserID == "" {
			skipped++
			continue
		}

		oldLevel := oldCalcLevel(entry.OldXP)
		newXP := newTotalXPForLevel(oldLevel)

		uCtx, uCancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := coll.UpdateOne(uCtx,
			bson.M{"user_id": entry.UserID},
			bson.M{"$set": bson.M{"user_id": entry.UserID, "xp": newXP}},
			options.UpdateOne().SetUpsert(true),
		)
		uCancel()

		if err != nil {
			log.Printf("  skip %s (db error): %v", entry.UserID, err)
			skipped++
			continue
		}

		fmt.Printf("  ✓  %-22s  old xp %-8d → level %-4d → new xp %d\n",
			entry.UserID, entry.OldXP, oldLevel, newXP)
		imported++
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("scan: %v", err)
	}

	fmt.Printf("\nDone — %d migrated, %d skipped\n", imported, skipped)
}
