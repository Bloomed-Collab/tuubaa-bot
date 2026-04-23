package config

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/S42yt/tuubaa-bot/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type guildConfig struct {
	GuildID        string            `bson:"guild_id"`
	Roles          map[string]string `bson:"roles"`
	LevelRoles     map[string]string `bson:"level_roles"`
	WelcomeChannel string            `bson:"welcome_channel"`
	MainChannel    string            `bson:"main_channel"`
	CounterChannel string            `bson:"counter_channel"`
	LogsChannel    string            `bson:"logs_channel"`
	BotChannel     string            `bson:"bot_channel"`
}

type channelCacheEntry struct {
	channelID string
	expiresAt time.Time
}

var (
	channelCacheMu sync.RWMutex
	channelCache   = map[string]channelCacheEntry{}
)

func cacheKey(guildID, key string) string { return guildID + "|" + key }

func GetChannelCached(guildID, key string) (string, error) {
	k := cacheKey(guildID, key)
	now := time.Now()

	channelCacheMu.RLock()
	if e, ok := channelCache[k]; ok && now.Before(e.expiresAt) {
		channelCacheMu.RUnlock()
		return e.channelID, nil
	}
	channelCacheMu.RUnlock()

	ch, err := GetChannel(guildID, key)
	if err != nil {
		return ch, err
	}

	channelCacheMu.Lock()
	channelCache[k] = channelCacheEntry{channelID: ch, expiresAt: now.Add(5 * time.Minute)}
	channelCacheMu.Unlock()

	return ch, nil
}

func GetRole(guildID, key string) (string, error) {
	db := core.NewMongoHandler()
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	if err := db.Connect(ctx); err != nil {
		return "", err
	}
	defer db.Disconnect(ctx)

	coll := db.Collection("guild_configs")
	var cfg guildConfig
	if err := coll.FindOne(ctx, bson.M{"guild_id": guildID}).Decode(&cfg); err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}
	if cfg.Roles == nil {
		return "", nil
	}
	if v, ok := cfg.Roles[key]; ok {
		return v, nil
	}
	return "", nil
}

func GetRoles(guildID string) (map[string]string, error) {
	db := core.NewMongoHandler()
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	if err := db.Connect(ctx); err != nil {
		return nil, err
	}
	defer db.Disconnect(ctx)

	coll := db.Collection("guild_configs")
	var cfg guildConfig
	if err := coll.FindOne(ctx, bson.M{"guild_id": guildID}).Decode(&cfg); err != nil {
		if err == mongo.ErrNoDocuments {
			return map[string]string{}, nil
		}
		return nil, err
	}
	if cfg.Roles == nil {
		return map[string]string{}, nil
	}
	return cfg.Roles, nil
}

func GetChannel(guildID, key string) (string, error) {
	db := core.NewMongoHandler()
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	if err := db.Connect(ctx); err != nil {
		return "", err
	}
	defer db.Disconnect(ctx)

	coll := db.Collection("guild_configs")
	var cfg guildConfig
	if err := coll.FindOne(ctx, bson.M{"guild_id": guildID}).Decode(&cfg); err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}
	switch key {
	case "welcome":
		return cfg.WelcomeChannel, nil
	case "main":
		return cfg.MainChannel, nil
	case "counterchannel":
		return cfg.CounterChannel, nil
	case "logs":
		return cfg.LogsChannel, nil
	case "bot":
		return cfg.BotChannel, nil
	}
	return "", nil
}

func GetLevelRole(guildID string, level int) (string, error) {
	db := core.NewMongoHandler()
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	if err := db.Connect(ctx); err != nil {
		return "", err
	}
	defer db.Disconnect(ctx)

	coll := db.Collection("guild_configs")
	var cfg guildConfig
	if err := coll.FindOne(ctx, bson.M{"guild_id": guildID}).Decode(&cfg); err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}
	if cfg.LevelRoles == nil {
		return "", nil
	}
	if v, ok := cfg.LevelRoles[strconv.Itoa(level)]; ok {
		return v, nil
	}
	return "", nil
}
