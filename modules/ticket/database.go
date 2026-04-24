package ticket

import (
	"context"
	"time"

	"github.com/S42yt/tuubaa-bot/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const colTickets = "tickets"

type ticketEntry struct {
	GuildID   string `bson:"guild_id"`
	ChannelID string `bson:"channel_id"`
	MessageID string `bson:"message_id"`
	UserID    string `bson:"user_id"`
	Kind      string `bson:"kind"`
	ClaimedBy string `bson:"claimed_by,omitempty"`
}

func saveTicket(t ticketEntry) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := core.DB().Collection(colTickets).UpdateOne(ctx,
		bson.M{"guild_id": t.GuildID, "channel_id": t.ChannelID},
		bson.M{"$set": t},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func getTicket(guildID, channelID string) (*ticketEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var e ticketEntry
	err := core.DB().Collection(colTickets).FindOne(ctx, bson.M{
		"guild_id":   guildID,
		"channel_id": channelID,
	}).Decode(&e)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &e, err
}

func claimTicket(guildID, channelID, claimedBy string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := core.DB().Collection(colTickets).UpdateOne(ctx,
		bson.M{"guild_id": guildID, "channel_id": channelID},
		bson.M{"$set": bson.M{"claimed_by": claimedBy}},
	)
	return err
}

func deleteTicket(guildID, channelID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	core.DB().Collection(colTickets).DeleteOne(ctx, bson.M{ 
		"guild_id":   guildID,
		"channel_id": channelID,
	})
}
