package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/S42yt/tuubaa-bot/core"
	logger "github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	lastCounterUpdate = map[string]time.Time{}
	lastMu            sync.Mutex
	counterInterval   = 5 * time.Minute
)

func init() {
	core.On(guildMemberAdd)
	core.On(guildMemberRemove)
}

func guildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	go handleCounterEvent(s, m.GuildID)
}

func guildMemberRemove(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	go handleCounterEvent(s, m.GuildID)
}

func handleCounterEvent(s *discordgo.Session, guildID string) {
	lastMu.Lock()
	if t, ok := lastCounterUpdate[guildID]; ok {
		if time.Since(t) < counterInterval {
			lastMu.Unlock()
			logger.Debug("counter_events: skipping update for %s (rate limited)", guildID)
			return
		}
	}
	lastCounterUpdate[guildID] = time.Now()
	lastMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	var doc bson.M
	if err := core.DB().Collection("guild_configs").FindOne(ctx, bson.M{"guild_id": guildID}).Decode(&doc); err != nil {
		logger.Debug("counter_events: no config for guild %s: %v", guildID, err)
		return
	}

	chID, _ := doc["counter_channel"].(string)
	if chID == "" {
		logger.Debug("counter_events: counter_channel not configured for %s", guildID)
		return
	}

	var memberCount int
	if g, err := s.GuildWithCounts(guildID); err == nil && g != nil {
		memberCount = g.ApproximateMemberCount
	}

	newName := fmt.Sprintf("「👥」Kinder✩%d", memberCount)
	if _, err := s.ChannelEdit(chID, &discordgo.ChannelEdit{Name: newName}); err != nil {
		logger.Warn("counter_events: failed to update channel %s: %v", chID, err)
		return
	}

	logger.Debug("counter_events: updated counter for guild %s -> %s", guildID, newName)
}
