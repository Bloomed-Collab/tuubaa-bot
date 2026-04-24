package level

import (
	"sync"
	"time"

	"github.com/S42yt/tuubaa-bot/core"
	"github.com/bwmarrin/discordgo"
)

var (
	vMu sync.RWMutex
	vChannelByUser = map[string]string{}
	vGuildByUser   = map[string]string{}

	vStopMu  sync.Mutex
	vStopMap = map[string]chan struct{}{}
)

func init() {
	core.On(voiceStateHandler)
}

func voiceStateHandler(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	if v.VoiceState == nil {
		return
	}

	userID := v.UserID
	guildID := v.GuildID

	if m, err := s.State.Member(guildID, userID); err == nil && m.User != nil && m.User.Bot {
		return
	}

	if v.ChannelID == "" {
		stopVoiceTicker(userID)
		return
	}

	vMu.Lock()
	vChannelByUser[userID] = v.ChannelID
	vGuildByUser[userID] = guildID
	vMu.Unlock()

	vStopMu.Lock()
	if _, running := vStopMap[userID]; !running {
		stop := make(chan struct{})
		vStopMap[userID] = stop
		go voiceTicker(s, userID, guildID, stop)
	}
	vStopMu.Unlock()
}

func voiceTicker(s *discordgo.Session, userID, guildID string, stop chan struct{}) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			h, m, _ := time.Now().Clock()
			if isLockdown(h*60 + m) {
				continue
			}

			if eligible, count := eligibleVoiceCount(s, guildID, userID); eligible {
				go addXP(s, guildID, userID, voiceXPForCount(count))
			}
		}
	}
}

func stopVoiceTicker(userID string) {
	vStopMu.Lock()
	if stop, ok := vStopMap[userID]; ok {
		close(stop)
		delete(vStopMap, userID)
	}
	vStopMu.Unlock()

	vMu.Lock()
	delete(vChannelByUser, userID)
	delete(vGuildByUser, userID)
	vMu.Unlock()
}

func eligibleVoiceCount(s *discordgo.Session, guildID, userID string) (bool, int) {
	vMu.RLock()
	channelID := vChannelByUser[userID]
	vMu.RUnlock()
	if channelID == "" {
		return false, 0
	}

	guild, err := s.State.Guild(guildID)
	if err != nil || guild == nil {
		return false, 0
	}

	var userVS *discordgo.VoiceState
	eligibleInChannel := 0
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID != channelID {
			continue
		}

		mem, merr := s.State.Member(guildID, vs.UserID)
		if merr != nil || mem == nil || mem.User == nil || mem.User.Bot {
			continue
		}

		if isVoiceEligible(vs) {
			eligibleInChannel++
		}
		if vs.UserID == userID {
			userVS = vs
		}
	}

	return userVS != nil && isVoiceEligible(userVS) && eligibleInChannel >= 2, eligibleInChannel
}

func isVoiceEligible(vs *discordgo.VoiceState) bool {
	if vs == nil {
		return false
	}
	if vs.Deaf || vs.SelfDeaf {
		return false
	}
	if vs.Mute || vs.SelfMute {
		return false
	}
	return true
}
