package events

import (
	"context"
	"fmt"
	"time"

	"github.com/S42yt/tuubaa-bot/core"
	vembed "github.com/S42yt/tuubaa-bot/modules/misc/embed"
	logger "github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
)

func init() {
	core.On(welcomeHandler)
}

func welcomeHandler(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	logger.Debug("welcomeHandler: join from user %s in guild %s", m.User.ID, m.GuildID)

	const minAge = 60 * 24 * time.Hour // 2 months
	createdAt, err := discordgo.SnowflakeTimestamp(m.User.ID)
	if err != nil {
		logger.Warn("welcomeHandler: failed to parse snowflake for %s: %v", m.User.ID, err)
	} else if time.Since(createdAt) < minAge {
		ageDays := int(time.Since(createdAt).Hours() / 24)
		logger.Debug("welcomeHandler: account %s too new (%d days), sending DM and skipping welcome", m.User.ID, ageDays)
		if ch, dmErr := s.UserChannelCreate(m.User.ID); dmErr != nil {
			logger.Warn("welcomeHandler: failed to create DM channel for %s: %v", m.User.ID, dmErr)
		} else if _, sendErr := s.ChannelMessageSend(ch.ID, "Dein Discord Account ist leider noch nicht alt genug, um dem Server beizutreten. Bitte versuche es erneut, wenn dein Account mindestens 2 Monate alt ist."); sendErr != nil {
			logger.Warn("welcomeHandler: failed to send DM to %s: %v", m.User.ID, sendErr)
		}
		return
	}

	time.Sleep(5 * time.Second)

	if _, err := s.GuildMember(m.GuildID, m.User.ID); err != nil {
		logger.Debug("welcomeHandler: member %s no longer in guild after sleep, skipping welcome", m.User.ID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var doc struct {
		WelcomeChannel string `bson:"welcome_channel"`
		MainChannel    string `bson:"main_channel"`
	}
	if err := core.DB().Collection("guild_configs").FindOne(ctx, map[string]interface{}{"guild_id": m.GuildID}).Decode(&doc); err != nil {
		logger.Debug("welcomeHandler: no config found for guild %s: %v", m.GuildID, err)
		return
	}
	if doc.WelcomeChannel == "" {
		logger.Debug("welcomeHandler: welcome_channel empty for guild %s", m.GuildID)
		return
	}

	avatarURL := m.User.AvatarURL("1024")
	displayName := m.User.DisplayName()
	memberId := m.User.ID

	var memberCount int
	if g, err := s.GuildWithCounts(m.GuildID); err == nil && g != nil {
		memberCount = g.ApproximateMemberCount
	}

	buf, err := vembed.BuildWelcomeImage(avatarURL, displayName, memberCount)
	if err != nil {
		logger.Warn("welcomeHandler: build image failed: %v", err)
		_, _ = s.ChannelMessageSend(doc.WelcomeChannel, fmt.Sprintf("Welcome image generation failed: %v", err))
		return
	}

	file := &discordgo.File{Name: "welcome.png", ContentType: "image/png", Reader: buf}

	comps, cerr := vembed.BuildWelcomeComponents(avatarURL, doc.MainChannel, displayName, memberCount, memberId)
	if cerr != nil {
		logger.Debug("welcomeHandler: build components failed: %v", cerr)
	}

	msg := &discordgo.MessageSend{
		Files: []*discordgo.File{file},
	}
	if cerr == nil && len(comps) > 0 {
		msg.Components = comps
		msg.Flags = discordgo.MessageFlagsIsComponentsV2 | discordgo.MessageFlagsSuppressNotifications
	}

	sent, err := s.ChannelMessageSendComplex(doc.WelcomeChannel, msg)
	if err != nil {
		logger.Error("welcomeHandler: send failed to %s: %v", doc.WelcomeChannel, err)
		return
	}
	logger.Debug("welcomeHandler: sent welcome message %s to %s", sent.ID, doc.WelcomeChannel)

	if doc.MainChannel != "" {
		content := fmt.Sprintf("Ein neuer gefangener im **goldenen Van**! Heißt <@%s> willkommen <:HeyTuba:1137369596496707624>", m.User.ID)
		if _, err := s.ChannelMessageSend(doc.MainChannel, content); err != nil {
			logger.Warn("welcomeHandler: send main channel failed to %s: %v", doc.MainChannel, err)
		} else {
			logger.Debug("welcomeHandler: sent main announcement to %s", doc.MainChannel)
		}
	}
}
