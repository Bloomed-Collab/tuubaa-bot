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

func handleSetLevelRole(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if err := deferConfigResponse(s, i); err != nil {
		return err
	}

	data := i.ApplicationCommandData().Options[0]
	var levelStr string
	var targetRoleID string

	for _, opt := range data.Options {
		switch opt.Name {
		case "level":
			levelStr = opt.StringValue()
		case "role":
			if r := opt.RoleValue(s, i.GuildID); r != nil {
				targetRoleID = r.ID
			} else {
				targetRoleID = opt.StringValue()
			}
		}
	}

	if levelStr == "" || targetRoleID == "" {
		return respond(s, i, "Ungültige Argumente")
	}

	db := core.NewMongoHandler()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.Connect(ctx); err != nil {
		return respond(s, i, fmt.Sprintf("DB-Verbindung fehlgeschlagen: %v", err))
	}
	defer db.Disconnect(ctx)

	coll := db.Collection("guild_configs")
	filter := bson.M{"guild_id": i.GuildID}
	update := bson.M{"$set": bson.M{fmt.Sprintf("level_roles.%s", levelStr): targetRoleID}}
	res, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return respond(s, i, fmt.Sprintf("Speichern fehlgeschlagen: %v", err))
	}
	if res.MatchedCount == 0 {
		doc := bson.M{"guild_id": i.GuildID, "level_roles": bson.M{levelStr: targetRoleID}}
		if _, err := coll.InsertOne(ctx, doc); err != nil {
			return respond(s, i, fmt.Sprintf("Erstellen fehlgeschlagen: %v", err))
		}
	}

	resp := vembed.BuildLevelRoleSetResponse(levelStr, targetRoleID, i.Member.User.Username)
	return editResponseData(s, i, resp)
}
