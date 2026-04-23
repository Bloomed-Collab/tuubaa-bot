package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/S42yt/tuubaa-bot/core"
	vembed "github.com/S42yt/tuubaa-bot/modules/config/embed"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func handleSetRole(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if err := deferConfigResponse(s, i); err != nil {
		return err
	}

	data := i.ApplicationCommandData().Options[0]
	var roleKey string
	var targetRoleID string

	for _, opt := range data.Options {
		switch opt.Name {
		case "role":
			roleKey = opt.StringValue()
		case "target":
			if r := opt.RoleValue(s, i.GuildID); r != nil {
				targetRoleID = r.ID
			} else {
				targetRoleID = opt.StringValue()
			}
		}
	}

	if roleKey == "" || targetRoleID == "" {
		return respond(s, i, "Invalid arguments")
	}

	db := core.NewMongoHandler()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.Connect(ctx); err != nil {
		return respond(s, i, fmt.Sprintf("Failed to connect to DB: %v", err))
	}
	defer db.Disconnect(ctx)

	coll := db.Collection("guild_configs")
	filter := bson.M{"guild_id": i.GuildID}
	update := bson.M{"$set": bson.M{fmt.Sprintf("roles.%s", roleKey): targetRoleID}}
	res, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return respond(s, i, fmt.Sprintf("Failed to save config: %v", err))
	}
	if res.MatchedCount == 0 {
		doc := bson.M{"guild_id": i.GuildID, "roles": bson.M{roleKey: targetRoleID}}
		if _, err := coll.InsertOne(ctx, doc); err != nil {
			return respond(s, i, fmt.Sprintf("Failed to create config: %v", err))
		}
	}

	resp := vembed.BuildRoleSetResponse(roleKey, targetRoleID, i.Member.User.Username)
	return editResponseData(s, i, resp)
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) error {
	data := &discordgo.InteractionResponseData{Content: content, Flags: commandVisibilityFlags(i)}
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	})
	if err == nil {
		return nil
	}
	_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	return editErr
}

func deferConfigResponse(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

func editResponseData(s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.InteractionResponseData) error {
	params := &discordgo.WebhookParams{
		Flags: data.Flags | commandVisibilityFlags(i),
	}
	if len(data.Components) > 0 {
		params.Components = data.Components
	} else {
		params.Content = data.Content
	}
	_, err := s.FollowupMessageCreate(i.Interaction, true, params)
	if err != nil {
		fallback := "Konfiguration gespeichert, aber Antwort konnte nicht aktualisiert werden."
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &fallback})
		return err
	}
	return err
}

func commandVisibilityFlags(i *discordgo.InteractionCreate) discordgo.MessageFlags {
	if i == nil || i.GuildID == "" {
		return 0
	}
	botChannelID, err := getConfiguredBotChannel(i.GuildID)
	if err != nil || botChannelID == "" {
		return 0
	}
	if i.ChannelID != botChannelID {
		return discordgo.MessageFlagsEphemeral
	}
	return 0
}

func getConfiguredBotChannel(guildID string) (string, error) {
	db := core.NewMongoHandler()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Connect(ctx); err != nil {
		return "", err
	}
	defer db.Disconnect(ctx)

	var doc struct {
		BotChannel string `bson:"bot_channel"`
	}
	err := db.Collection("guild_configs").FindOne(ctx, bson.M{"guild_id": guildID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}
	return doc.BotChannel, nil
}
