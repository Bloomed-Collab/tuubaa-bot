package events

import (
	"context"
	"time"

	"github.com/S42yt/tuubaa-bot/core"
	"github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var goodnightLoopStarted bool

func init() {
	core.On(startGoodnightLoop)
}

func startGoodnightLoop(s *discordgo.Session, r *discordgo.Ready) {
	if goodnightLoopStarted {
		return
	}
	goodnightLoopStarted = true

	go func() {
		loc, err := time.LoadLocation("Europe/Berlin")
		if err != nil {
			logger.Error("startGoodnightLoop: failed to load timezone Europe/Berlin: %v", err)
			return
		}

		for {
			now := time.Now().In(loc)

			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
			duration := next.Sub(now)

			logger.Info("startGoodnightLoop: next goodnight message in %v", duration)
			time.Sleep(duration)

			sendGoodnightMessages(s)
		}
	}()
}

func sendGoodnightMessages(s *discordgo.Session) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := core.DB().Collection("guild_configs").Find(ctx, bson.M{})
	if err != nil {
		logger.Error("sendGoodnightMessages: failed to find guild_configs: %v", err)
		return
	}
	defer cursor.Close(ctx)

	type GuildConfig struct {
		MainChannel string `bson:"main_channel"`
	}

	var configs []GuildConfig
	if err := cursor.All(ctx, &configs); err != nil {
		logger.Error("sendGoodnightMessages: failed to decode guild_configs: %v", err)
		return
	}

	for _, cfg := range configs {
		if cfg.MainChannel == "" {
			continue
		}

		_, err := s.ChannelMessageSend(cfg.MainChannel, "Gute Nacht, Gefangene des Vans! <:TuubaaAwake:1244353418894643271>")
		if err != nil {
			logger.Warn("sendGoodnightMessages: failed to send to channel %s: %v", cfg.MainChannel, err)
		}
	}
}
