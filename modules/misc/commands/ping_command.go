package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/S42yt/tuubaa-bot/core"
	v2 "github.com/S42yt/tuubaa-bot/utils/embed"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func PingHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		start := time.Now()
		_, chErr := s.Channel(i.ChannelID)
		botLatency := time.Since(start)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		startDB := time.Now()
		dbErr := core.DB().RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err()
		dbLatency := time.Since(startDB)

		title := v2.NewTextDisplayBuilder().SetContent("### Pong!").Build()

		botLine := fmt.Sprintf("Bot: %dms", botLatency.Milliseconds())
		if chErr != nil {
			botLine = fmt.Sprintf("Bot: error (%v)", chErr)
		}

		dbLine := fmt.Sprintf("DB: %dms", dbLatency.Milliseconds())
		if dbErr != nil {
			dbLine = fmt.Sprintf("DB: error (%v)", dbErr)
		}

		body := v2.NewTextDisplayBuilder().SetContent(botLine + "\n" + dbLine).Build()

		accent := 0x2ecc71
		if chErr != nil || dbErr != nil {
			accent = 0x992222
		}

		comp := v2.NewContainerBuilder().SetAccentColor(accent).
			AddComponent(title).
			AddComponent(body).
			Build()

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Components: []discordgo.MessageComponent{comp},
				Flags:      discordgo.MessageFlagsIsComponentsV2,
			},
		})
	}
}
