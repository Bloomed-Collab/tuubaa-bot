package gallery

import (
	"context"
	"time"

	"github.com/S42yt/tuubaa-bot/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	colThreads = "gallery_threads"
	colPosts   = "gallery_posts"
)

type threadEntry struct {
	GuildID  string `bson:"guild_id"`
	UserID   string `bson:"user_id"`
	ThreadID string `bson:"thread_id"`
}

type postEntry struct {
	GuildID   string `bson:"guild_id"`
	ChannelID string `bson:"channel_id"`
	MessageID string `bson:"message_id"`
	ThreadID  string `bson:"thread_id"`
	PostID    string `bson:"post_id"`
}

func getThread(guildID, userID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var entry threadEntry
	err := core.DB().Collection(colThreads).FindOne(ctx, bson.M{
		"guild_id": guildID,
		"user_id":  userID,
	}).Decode(&entry)
	if err == mongo.ErrNoDocuments {
		return "", nil
	}
	return entry.ThreadID, err
}

func saveThread(guildID, userID, threadID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := core.DB().Collection(colThreads).UpdateOne(ctx,
		bson.M{"guild_id": guildID, "user_id": userID},
		bson.M{"$set": bson.M{"guild_id": guildID, "user_id": userID, "thread_id": threadID}},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func getPost(guildID, channelID, messageID string) (*postEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var entry postEntry
	err := core.DB().Collection(colPosts).FindOne(ctx, bson.M{
		"guild_id":   guildID,
		"channel_id": channelID,
		"message_id": messageID,
	}).Decode(&entry)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func savePost(guildID, channelID, messageID, threadID, postID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := core.DB().Collection(colPosts).UpdateOne(ctx,
		bson.M{"guild_id": guildID, "channel_id": channelID, "message_id": messageID},
		bson.M{"$set": postEntry{
			GuildID:   guildID,
			ChannelID: channelID,
			MessageID: messageID,
			ThreadID:  threadID,
			PostID:    postID,
		}},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func deletePost(guildID, channelID, messageID string) (threadID, postID string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var entry postEntry
	err = core.DB().Collection(colPosts).FindOneAndDelete(ctx, bson.M{
		"guild_id":   guildID,
		"channel_id": channelID,
		"message_id": messageID,
	}).Decode(&entry)
	if err == mongo.ErrNoDocuments {
		return "", "", nil
	}
	if err != nil {
		return "", "", err
	}
	return entry.ThreadID, entry.PostID, nil
}
