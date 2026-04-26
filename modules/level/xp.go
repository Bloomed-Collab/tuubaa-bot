package level

import (
	"math"
	"sync"
	"time"

	cfg "github.com/S42yt/tuubaa-bot/modules/config"
	ulog "github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
)

var levelRoleThresholds = []int{20, 40, 60, 80, 100}

var (
	hourlyMu     sync.Mutex
	hourlyHour   time.Time
	hourlyEarned = map[string]int64{}

	dailyMu     sync.Mutex
	dailyDate   time.Time
	dailyEarned = map[string]int64{}
)

func clampToHourlyLimit(userID string, toAdd int64) int64 {
	hourlyMu.Lock()
	defer hourlyMu.Unlock()
	thisHour := time.Now().Truncate(time.Hour)
	if !hourlyHour.Equal(thisHour) {
		hourlyHour = thisHour
		hourlyEarned = map[string]int64{}
	}
	earned := hourlyEarned[userID]
	if earned >= hourlyXPLimit {
		return 0
	}
	if remaining := hourlyXPLimit - earned; toAdd > remaining {
		toAdd = remaining
	}
	hourlyEarned[userID] += toAdd
	return toAdd
}

func clampToDailyLimit(userID string, toAdd int64) int64 {
	dailyMu.Lock()
	defer dailyMu.Unlock()
	today := time.Now().Truncate(24 * time.Hour)
	if !dailyDate.Equal(today) {
		dailyDate = today
		dailyEarned = map[string]int64{}
	}
	earned := dailyEarned[userID]
	if earned >= dailyXPLimit {
		return 0
	}
	if remaining := dailyXPLimit - earned; toAdd > remaining {
		toAdd = remaining
	}
	dailyEarned[userID] += toAdd
	return toAdd
}

func addXP(s *discordgo.Session, guildID, userID string, amount float64) {
	if amount <= 0 {
		return
	}
	toAdd := int64(math.Floor(amount))
	if toAdd == 0 {
		return
	}
	toAdd = clampToHourlyLimit(userID, toAdd)
	if toAdd == 0 {
		return
	}
	toAdd = clampToDailyLimit(userID, toAdd)
	if toAdd == 0 {
		return
	}

	current, err := getXP(userID)
	if err != nil {
		ulog.Warn("level: getXP(%s): %v", userID, err)
		return
	}

	prevLevel := calcLevel(current)
	newXP := current + toAdd
	nextThreshold := totalXPForLevel(prevLevel + 1)
	if prevLevel < lvlMax && current < nextThreshold && newXP >= nextThreshold {
		newXP = nextThreshold
	}

	newLevel := calcLevel(newXP)
	levelUpXP := xpToNextLevel(current)

	if err := upsertXP(userID, newXP); err != nil {
		ulog.Warn("level: upsertXP(%s): %v", userID, err)
		return
	}

	if newLevel > prevLevel && levelUpXP > 0 {
		go sendLevelUp(s, guildID, userID, newLevel)
	}
}

func sendLevelUp(s *discordgo.Session, guildID, userID string, newLevel int) {
	channelID, err := cfg.GetChannel(guildID, "bot")
	if err != nil || channelID == "" {
		return
	}

	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		ulog.Warn("level: GuildMember(%s): %v", userID, err)
		return
	}

	displayName := ""
	if member.User != nil {
		displayName = member.User.GlobalName
		if displayName == "" {
			displayName = member.User.Username
		}
	}
	if member.Nick != "" {
		displayName = member.Nick
	}

	var assignedRoleName string
	if roleID, roleErr := cfg.GetLevelRole(guildID, newLevel); roleErr == nil && roleID != "" {
		for _, threshold := range levelRoleThresholds {
			oldRoleID, oldRoleErr := cfg.GetLevelRole(guildID, threshold)
			if oldRoleErr == nil && oldRoleID != "" {
				_ = s.GuildMemberRoleRemove(guildID, userID, oldRoleID)
			}
		}
		if err := s.GuildMemberRoleAdd(guildID, userID, roleID); err == nil {
			if role, rErr := s.State.Role(guildID, roleID); rErr == nil && role != nil {
				assignedRoleName = role.Name
			}
		}
	}

	comps := buildLevelUpComponents(displayName, newLevel, assignedRoleName)
	_, _ = s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Components: comps,
		Flags:      discordgo.MessageFlagsIsComponentsV2,
	})
}
