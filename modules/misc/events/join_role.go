package events

import (
	"github.com/S42yt/tuubaa-bot/core"
	cfg "github.com/S42yt/tuubaa-bot/modules/config"
	logger "github.com/S42yt/tuubaa-bot/utils/logger"
	"github.com/bwmarrin/discordgo"
)

func init() {
	core.On(joinRoleHandler)
}

func joinRoleHandler(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	roleID, err := cfg.GetRole(m.GuildID, "join_role")
	if err != nil || roleID == "" {
		return
	}

	for _, r := range m.Roles {
		if r == roleID {
			return
		}
	}

	if err := s.GuildMemberRoleAdd(m.GuildID, m.User.ID, roleID); err != nil {
		logger.Warn("joinRoleHandler: failed to assign join role to %s in %s: %v", m.User.ID, m.GuildID, err)
	}
}
