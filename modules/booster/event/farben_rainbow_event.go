package event

import (
	"context"
	"fmt"
	"time"

	"github.com/S42yt/tuubaa-bot/core"
	"github.com/S42yt/tuubaa-bot/modules/config"
	ulog "github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type RainbowUser struct {
	UserID       string    `bson:"user_id"`
	GuildID      string    `bson:"guild_id"`
	CurrentIndex int       `bson:"current_index"`
	Active       bool      `bson:"active"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

var SelectableRoles = []string{
	"Unschuldiges Kind",
	"Verdächtiges Kind",
	"Schuldiges Kind",
	"Mit Entführer",
	"Meisterentführer",
	"Beifahrer",
	"Van Upgrader",
}

var ChoiceKey = map[string]string{
	"Unschuldiges Kind": "ROLE_UNSCHULDIGES_KIND",
	"Verdächtiges Kind": "ROLE_VERDAECHTIGES_KIND",
	"Schuldiges Kind":   "ROLE_SCHULDIGES_KIND",
	"Mit Entführer":     "ROLE_MIT_ENTFUEHRER",
	"Meisterentführer":  "ROLE_MEISTERENTFUEHRER",
	"Beifahrer":         "ROLE_BEIFAHRER",
	"Van Upgrader":      "ROLE_VAN_UPGRADER",
}

func SetRainbowActive(userID, guildID string, active bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := core.DB().Collection("rainbow_users")
	filter := bson.M{"user_id": userID, "guild_id": guildID}
	update := bson.M{
		"$set": bson.M{
			"active":     active,
			"updated_at": time.Now(),
		},
	}
	if active {
		update["$setOnInsert"] = bson.M{"current_index": 0}
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := coll.UpdateOne(ctx, filter, update, opts)
	return err
}

func StartRainbowLoop(s *discordgo.Session) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			ulog.Debug("Rainbow loop: checking for users due for update")
			processRainbowUsers(s)
		}
	}()
}

func processRainbowUsers(s *discordgo.Session) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	coll := core.DB().Collection("rainbow_users")
	thirtyMinsAgo := time.Now().Add(-30 * time.Minute)
	filter := bson.M{
		"active": true,
		"$or": []bson.M{
			{"updated_at": bson.M{"$lte": thirtyMinsAgo}},
			{"updated_at": bson.M{"$exists": false}},
		},
	}

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		ulog.Error("Rainbow loop: failed to query users: %v", err)
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var user RainbowUser
		if err := cursor.Decode(&user); err != nil {
			ulog.Warn("Rainbow loop: failed to decode user: %v", err)
			continue
		}

		nextIndex := (user.CurrentIndex + 1) % len(SelectableRoles)

		err := updateMemberRole(s, user.GuildID, user.UserID, user.CurrentIndex, nextIndex)
		if err != nil {
			ulog.Warn("Rainbow loop: failed to update role for user %s in guild %s: %v", user.UserID, user.GuildID, err)
			_, _ = coll.UpdateOne(ctx, bson.M{"user_id": user.UserID, "guild_id": user.GuildID}, bson.M{
				"$set": bson.M{
					"updated_at": time.Now(),
				},
			})
			continue
		}

		_, _ = coll.UpdateOne(ctx, bson.M{"user_id": user.UserID, "guild_id": user.GuildID}, bson.M{
			"$set": bson.M{
				"current_index": nextIndex,
				"updated_at":    time.Now(),
			},
		})
		ulog.Debug("Rainbow loop: updated user %s to role %s", user.UserID, SelectableRoles[nextIndex])
	}
}

func updateMemberRole(s *discordgo.Session, guildID, userID string, oldIndex, newIndex int) error {
	rolesMap, err := config.GetRoles(guildID)
	if err != nil {
		return fmt.Errorf("failed to get roles: %w", err)
	}

	//oldRoleName := SelectableRoles[oldIndex]
	newRoleName := SelectableRoles[newIndex]

	//oldRoleKey := ChoiceKey[oldRoleName]
	newRoleKey := ChoiceKey[newRoleName]

	//oldRoleID := rolesMap[oldRoleKey]
	newRoleID := rolesMap[newRoleKey]

	if newRoleID == "" {
		return fmt.Errorf("new role %s not configured", newRoleName)
	}

	for _, name := range SelectableRoles {
		rid := rolesMap[ChoiceKey[name]]
		if rid == "" || rid == newRoleID {
			continue
		}
		
		_ = s.GuildMemberRoleRemove(guildID, userID, rid)
	}

	return s.GuildMemberRoleAdd(guildID, userID, newRoleID)
}
