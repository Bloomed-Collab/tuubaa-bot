package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/S42yt/tuubaa-bot/core"
	vembed "github.com/S42yt/tuubaa-bot/modules/config/embed"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func handleSetChannel(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if err := deferConfigResponse(s, i); err != nil {
		return err
	}

	data := i.ApplicationCommandData().Options[0]
	var targetChannelID string
	var whichKey string

	for _, opt := range data.Options {
		switch opt.Name {
		case "which":
			whichKey = opt.StringValue()
		case "channel":
			if c := opt.ChannelValue(s); c != nil {
				targetChannelID = c.ID
			} else {
				targetChannelID = opt.StringValue()
			}
		}
	}

	if whichKey == "" {
		return respond(s, i, "You must specify which config to set (e.g. welcome)")
	}
	if targetChannelID == "" {
		return respond(s, i, "Invalid channel provided")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := core.DB().Collection("guild_configs")
	filter := bson.M{"guild_id": i.GuildID}

	var fieldName string
	switch whichKey {
	case "welcome":
		fieldName = "welcome_channel"
	case "main":
		fieldName = "main_channel"
	case "counterchannel":
		fieldName = "counter_channel"
	case "logs":
		fieldName = "logs_channel"
	case "bot":
		fieldName = "bot_channel"
	default:
		return respond(s, i, "Unknown channel config key")
	}

	update := bson.M{"$set": bson.M{fieldName: targetChannelID}}
	res, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return respond(s, i, fmt.Sprintf("Failed to save config: %v", err))
	}
	if res.MatchedCount == 0 {
		doc := bson.M{"guild_id": i.GuildID}
		doc[fieldName] = targetChannelID
		if _, err := coll.InsertOne(ctx, doc); err != nil {
			return respond(s, i, fmt.Sprintf("Failed to create config: %v", err))
		}
	}

	resp := vembed.BuildChannelSetResponse(targetChannelID, i.Member.User.Username, whichKey)
	return editResponseData(s, i, resp)
}
