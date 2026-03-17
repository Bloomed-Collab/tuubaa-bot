package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/S42yt/tuubaa-bot/core"
	v2 "github.com/S42yt/tuubaa-bot/utils/embed"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type AIConfig struct {
	GuildID string `bson:"guild_id"`
	Enabled bool   `bson:"enabled"`
}

func AIToggleHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		guildID := i.GuildID
		if guildID == "" {
			return respondEphemeral(s, i, "This command can only be used in a guild")
		}

		db := core.NewMongoHandler()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.Connect(ctx); err != nil {
			return respondEphemeral(s, i, "Database error")
		}
		defer db.Disconnect(ctx)

		coll := db.Collection("ai_config")
		var config AIConfig
		err := coll.FindOne(ctx, bson.M{"guild_id": guildID}).Decode(&config)
		if err != nil && err != mongo.ErrNoDocuments {
			return respondEphemeral(s, i, "Error reading configuration")
		}

		newState := !config.Enabled
		opts := options.UpdateOne().SetUpsert(true)
		_, err = coll.UpdateOne(ctx, bson.M{"guild_id": guildID}, bson.M{
			"$set": bson.M{
				"guild_id": guildID,
				"enabled":  newState,
			},
		}, opts)

		if err != nil {
			return respondEphemeral(s, i, "Error updating configuration")
		}

		status := "🟢 **Aktiviert**"
		if !newState {
			status = "🔴 **Deaktiviert**"
		}

		title := v2.NewTextDisplayBuilder().
			SetContent(fmt.Sprintf("### KI %s", status)).
			Build()

		body := v2.NewTextDisplayBuilder().
			SetContent("Die KI-Antworten wurden " + map[bool]string{true: "aktiviert", false: "deaktiviert"}[newState]).
			Build()

		comp := v2.NewContainerBuilder().
			SetAccentColor(map[bool]int{true: 0x2ecc71, false: 0x992222}[newState]).
			AddComponent(title).
			AddComponent(body).
			Build()

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:      discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsIsComponentsV2,
				Components: []discordgo.MessageComponent{comp},
			},
		})
	}
}

func IsAIEnabled(guildID string) (bool, error) {
	db := core.NewMongoHandler()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Connect(ctx); err != nil {
		return false, err
	}
	defer db.Disconnect(ctx)

	coll := db.Collection("ai_config")
	var config AIConfig
	err := coll.FindOne(ctx, bson.M{"guild_id": guildID}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return true, nil
		}
		return false, err
	}

	return config.Enabled, nil
}

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsIsComponentsV2,
		},
	})
}
