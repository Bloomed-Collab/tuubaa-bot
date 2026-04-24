package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/S42yt/tuubaa-bot/core"
	_ "github.com/S42yt/tuubaa-bot/modules/booster"
	//_ "github.com/S42yt/tuubaa-bot/modules/chatgpt"
	_ "github.com/S42yt/tuubaa-bot/modules/config"
	_ "github.com/S42yt/tuubaa-bot/modules/level"
	_ "github.com/S42yt/tuubaa-bot/modules/misc"
	_ "github.com/S42yt/tuubaa-bot/modules/roleplay"
	_ "github.com/S42yt/tuubaa-bot/modules/ticket"
	logger "github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {

	debugLevel := false
	if os.Getenv("LOG_DEBUG") == "true" {
		debugLevel = true
	}

	logger.SetDebug(debugLevel)

	if err := godotenv.Load(); err != nil {
		logger.Warn(".env file not found, using system environment variables")
	}

	if err := core.InitDB(); err != nil {
		logger.Error("failed to connect to database: %v", err)
		os.Exit(2)
	}

	token := os.Getenv("TOKEN")
	if token == "" {
		token = os.Getenv("DISCORD_TOKEN")
	}

	if token == "" {
		logger.Error("Discord bot token not found. Set `DISCORD_TOKEN` or `TOKEN` environment variable, or create a `.env` file with TOKEN=your_token")
		os.Exit(2)
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		logger.Error("error creating Discord session: %v", err)
		os.Exit(2)
	}

	dg.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildPresences |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsMessageContent

	dg.AddHandlerOnce(ready)

	if err := dg.Open(); err != nil {
		logger.Error("error opening connection: %v", err)
		os.Exit(2)
	}
	defer dg.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = ctx
	_ = dg.UpdateStatusComplex(discordgo.UpdateStatusData{Status: "invisible"})
}

func ready(s *discordgo.Session, r *discordgo.Ready) {
	data := discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			{
				Name: "tuubaa :3",
				Type: discordgo.ActivityTypeWatching,
			},
		},
		Status: "online",
	}

	if err := s.UpdateStatusComplex(data); err != nil {
		logger.Error("failed to set presence: %v", err)
		return
	}

	if s.State != nil && s.State.User != nil {
		logger.Info("Logged in as %s#%s (%s)", s.State.User.Username, s.State.User.Discriminator, s.State.User.ID)
	}

	guildID := os.Getenv("GUILD_ID")
	if guildID == "" {
		if v, err := core.GetGuildIDCore("GUILD_ID"); err == nil {
			guildID = v
		}
	}

	if guildID != "" {
		if err := core.InitWithGuild(s, guildID); err != nil {
			logger.Error("failed to init command handler (guild): %v", err)
		} else {
			logger.Info("command handler initialized for guild %s", guildID)
		}
	} else {
		if err := core.Init(s); err != nil {
			logger.Error("failed to init command handler: %v", err)
		} else {
			logger.Info("command handler initialized")
		}
	}
}
