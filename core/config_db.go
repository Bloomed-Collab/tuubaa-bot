package core

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type guildConfig struct {
	GuildID string            `bson:"guild_id"`
	Roles   map[string]string `bson:"roles"`
}

func GetGuildIDCore(guildID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	var cfg guildConfig
	if err := DB().Collection("guild_configs").FindOne(ctx, bson.M{"guild_id": guildID}).Decode(&cfg); err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}

	return cfg.GuildID, nil
}
