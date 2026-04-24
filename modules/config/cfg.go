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
	GuildID             string            `bson:"guild_id"`
	Roles               map[string]string `bson:"roles"`
	LevelRoles          map[string]string `bson:"level_roles"`
	WelcomeChannel      string            `bson:"welcome_channel"`
	MainChannel         string            `bson:"main_channel"`
	CounterChannel      string            `bson:"counter_channel"`
	LogsChannel         string            `bson:"logs_channel"`
	BotChannel          string            `bson:"bot_channel"`
	GalleryForumChannel string            `bson:"gallery_forum_channel"`
	ArtChannel1         string            `bson:"art_channel_1"`
	ArtChannel2         string            `bson:"art_channel_2"`
	ArtChannel3         string            `bson:"art_channel_3"`
	TicketChannel       string            `bson:"ticket_channel"`
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
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	var cfg guildConfig
	if err := core.DB().Collection("guild_configs").FindOne(ctx, bson.M{"guild_id": guildID}).Decode(&cfg); err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}
	if v, ok := cfg.Roles[key]; ok {
		return v, nil
	}
	return "", nil
}

func GetRoles(guildID string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	var cfg guildConfig
	if err := core.DB().Collection("guild_configs").FindOne(ctx, bson.M{"guild_id": guildID}).Decode(&cfg); err != nil {
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
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	var cfg guildConfig
	if err := core.DB().Collection("guild_configs").FindOne(ctx, bson.M{"guild_id": guildID}).Decode(&cfg); err != nil {
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
	case "gallery_forum":
		return cfg.GalleryForumChannel, nil
	case "art_1":
		return cfg.ArtChannel1, nil
	case "art_2":
		return cfg.ArtChannel2, nil
	case "art_3":
		return cfg.ArtChannel3, nil
	case "ticket":
		return cfg.TicketChannel, nil
	}
	return "", nil
}

func GetArtChannels(guildID string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	var c guildConfig
	if err := core.DB().Collection("guild_configs").FindOne(ctx, bson.M{"guild_id": guildID}).Decode(&c); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	var channels []string
	for _, ch := range []string{c.ArtChannel1, c.ArtChannel2, c.ArtChannel3} {
		if ch != "" {
			channels = append(channels, ch)
		}
	}
	return channels, nil
}

func GetLevelRole(guildID string, level int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	var cfg guildConfig
	if err := core.DB().Collection("guild_configs").FindOne(ctx, bson.M{"guild_id": guildID}).Decode(&cfg); err != nil {
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
