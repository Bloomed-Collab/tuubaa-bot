package core

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	ulog "github.com/S42yt/tuubaa-bot/utils/logger"
	users "github.com/S42yt/tuubaa-bot/utils/users"
	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Name            string
	Description     string
	Options         []*discordgo.ApplicationCommandOption
	Handler         func(s *discordgo.Session, i *discordgo.InteractionCreate) error
	Cooldown        int
	OnlyBotChannel  bool
	DisabledUsers   map[string]bool
	DisabledChannel map[string]bool
	AllowAdmin      bool
	AllowStaff      bool
	AllowEveryone   bool
}

var (
	mu                         sync.RWMutex
	commands                   = map[string]*Command{}
	attachMu                   sync.Mutex
	interactionHandlerAttached bool
	eventsAttached             bool
	messageHandlerAttached     bool
)

func Register(cmd *Command) error {
	if cmd == nil || cmd.Name == "" {
		return errors.New("invalid command")
	}
	mu.Lock()
	defer mu.Unlock()
	if _, ok := commands[cmd.Name]; ok {
		return fmt.Errorf("command %s already registered", cmd.Name)
	}
	commands[cmd.Name] = cmd
	return nil
}

func GetAllCommands() map[string]*Command {
	mu.RLock()
	defer mu.RUnlock()

	result := make(map[string]*Command, len(commands))
	for name, cmd := range commands {
		result[name] = cmd
	}
	return result
}

func GetCommandCount() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(commands)
}

func Init(s *discordgo.Session) error {
	if s.State == nil || s.State.User == nil {
		return errors.New("session state not ready; call Init from Ready handler")
	}
	for _, g := range s.State.Guilds {
		if err := InitWithGuild(s, g.ID); err != nil {
			ulog.Warn("failed to init commands for guild %s: %v", g.ID, err)
		}
	}
	return nil
}

func InitWithGuild(s *discordgo.Session, guildID string) error {
	appID := s.State.User.ID

	mu.RLock()
	defer mu.RUnlock()

	if err := removeGlobalConflicts(s); err != nil {
		ulog.Warn("failed to remove global conflicting commands: %v", err)
	}

	if !modalHandlerAttached {
		s.AddHandler(modalInteractionHandler)
		modalHandlerAttached = true
	}

	var appCommands []*discordgo.ApplicationCommand
	for _, c := range commands {
		appCommands = append(appCommands, &discordgo.ApplicationCommand{
			Name:        c.Name,
			Description: c.Description,
			Options:     c.Options,
		})
	}

	if _, err := s.ApplicationCommandBulkOverwrite(appID, guildID, appCommands); err != nil {
		return err
	}

	ulog.Info("published guild commands for guild %s", guildID)

	attachMu.Lock()
	defer attachMu.Unlock()

	if !interactionHandlerAttached {
		s.AddHandler(interactionHandler)
		interactionHandlerAttached = true
	}

	if !messageHandlerAttached {
		s.AddHandler(messageHandler)
		messageHandlerAttached = true
	}

	if !eventsAttached {
		for _, h := range eventHandlers {
			s.AddHandler(h)
		}
		eventsAttached = true
	}

	return nil
}

func applyCommandPermissions(s *discordgo.Session, appID, guildID string, cmdIDMap map[string]string) error {
	mu.RLock()
	defer mu.RUnlock()

	for _, cmd := range commands {
		cmdID, ok := cmdIDMap[cmd.Name]
		if !ok {
			continue
		}

		var perms []*discordgo.ApplicationCommandPermissions

		if cmd.AllowEveryone {
			perms = append(perms, &discordgo.ApplicationCommandPermissions{
				ID:         guildID,
				Type:       discordgo.ApplicationCommandPermissionTypeRole,
				Permission: true,
			})
		} else {
			for _, u := range users.Users {
				perms = append(perms, &discordgo.ApplicationCommandPermissions{
					ID:         u.UserID,
					Type:       discordgo.ApplicationCommandPermissionTypeUser,
					Permission: true,
				})
			}

			if cmd.AllowAdmin || cmd.AllowStaff {
				roles, err := s.GuildRoles(guildID)
				if err != nil {
					ulog.Warn("could not fetch roles for guild %s: %v", guildID, err)
				} else {
					for _, role := range roles {
						if role.Permissions&discordgo.PermissionAdministrator != 0 {
							perms = append(perms, &discordgo.ApplicationCommandPermissions{
								ID:         role.ID,
								Type:       discordgo.ApplicationCommandPermissionTypeRole,
								Permission: true,
							})
						}
					}
				}
			}
		}

		if len(perms) == 0 {
			continue
		}

		err := s.ApplicationCommandPermissionsEdit(appID, guildID, cmdID, &discordgo.ApplicationCommandPermissionsList{
			Permissions: perms,
		})
		if err != nil {
			ulog.Warn("failed to set permissions for command %s: %v", cmd.Name, err)
		} else {
			ulog.Debug("set permissions for command %s (%s)", cmd.Name, cmdID)
		}
	}

	return nil
}

func removeGlobalConflicts(s *discordgo.Session) error {
	if s.State == nil || s.State.User == nil {
		return errors.New("session state not ready; call removeGlobalConflicts after Ready")
	}
	appID := s.State.User.ID

	globals, err := s.ApplicationCommands(appID, "")
	if err != nil {
		return fmt.Errorf("fetch global commands: %w", err)
	}

	mu.RLock()
	defer mu.RUnlock()

	for _, g := range globals {
		if _, ok := commands[g.Name]; ok {
			if err := s.ApplicationCommandDelete(appID, "", g.ID); err != nil {
				ulog.Warn("failed to delete global command %s (%s): %v", g.Name, g.ID, err)
			} else {
				ulog.Info("deleted global command %s (%s)", g.Name, g.ID)
			}
		}
	}

	return nil
}

func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	name := i.ApplicationCommandData().Name

	mu.RLock()
	cmd, ok := commands[name]
	mu.RUnlock()
	if !ok {
		_ = respondEphemeral(s, i, "This command no longer exists")
		return
	}

	if i.Member == nil {
		_ = respondEphemeral(s, i, "Could not validate member")
		return
	}

	member := i.Member.User

	if cmd.Cooldown > 0 {
		if remaining := checkCooldown(member.ID, cmd.Name); remaining > 0 {
			_ = respondEphemeral(s, i, fmt.Sprintf("You have a %d seconds cooldown", int(remaining)))
			return
		}
		setCooldown(member.ID, cmd.Name, cmd.Cooldown)
	}

	if cmd.OnlyBotChannel {
		if os.Getenv("BOT_CHANNEL") != "" && i.ChannelID != os.Getenv("BOT_CHANNEL") {
			_ = respondEphemeral(s, i, fmt.Sprintf("This command can only be executed in <#%s>", os.Getenv("BOT_CHANNEL")))
			return
		}
	}

	if cmd.DisabledUsers != nil {
		if cmd.DisabledUsers[member.ID] {
			_ = respondEphemeral(s, i, "You are excluded from this command")
			return
		}
	}

	if cmd.DisabledChannel != nil {
		if cmd.DisabledChannel[i.ChannelID] {
			_ = respondEphemeral(s, i, "This channel is excluded from this command")
			return
		}
	}

	isAdmin := i.Member.Permissions&discordgo.PermissionAdministrator != 0
	isAllowedUser := users.GetUserByID(member.ID) != nil

	if !(cmd.AllowEveryone || isAllowedUser || (cmd.AllowAdmin && isAdmin) || (cmd.AllowStaff && isAdmin)) {
		_ = respondEphemeral(s, i, "You do not have permission for this command")
		return
	}

	if err := cmd.Handler(s, i); err != nil {
		_ = respondEphemeral(s, i, "An internal error occurred")
	}
}

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) error {
	if i == nil || s == nil {
		return errors.New("nil session or interaction")
	}
	data := &discordgo.InteractionResponseData{Content: content, Flags: discordgo.MessageFlagsEphemeral}
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	})
}

var (
	cdMu    sync.Mutex
	cdStore = map[string]int64{}
)

func setCooldown(userID, cmdName string, seconds int) {
	cdMu.Lock()
	defer cdMu.Unlock()
	key := userID + ":" + cmdName
	cdStore[key] = time.Now().Unix() + int64(seconds)
}

func checkCooldown(userID, cmdName string) int64 {
	cdMu.Lock()
	defer cdMu.Unlock()
	key := userID + ":" + cmdName
	exp, ok := cdStore[key]
	if !ok {
		return 0
	}
	now := time.Now().Unix()
	if exp <= now {
		delete(cdStore, key)
		return 0
	}
	return exp - now
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if len(m.Content) == 0 || m.Content[0] != '!' {
		return
	}

	command := m.Content[1:]

	switch command {
	case "registerCommands":
		handleRegisterCommands(s, m)
	case "listCommands":
		handleListCommands(s, m)
	}
}

func handleRegisterCommands(s *discordgo.Session, m *discordgo.MessageCreate) {
	guildID := m.GuildID
	if guildID == "" {
		ch, err := s.Channel(m.ChannelID)
		if err != nil || ch.GuildID == "" {
			ulog.Warn("handleRegisterCommands: could not resolve guild ID: %v", err)
			s.ChannelMessageSend(m.ChannelID, "Could not resolve Guild ID")
			return
		}
		guildID = ch.GuildID
	}

	ulog.Debug("handleRegisterCommands: resolved GUILD_ID=%s", guildID)
	s.ChannelMessageSend(m.ChannelID, "Registering commands...")

	if err := InitWithGuild(s, guildID); err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to register commands: %v", err))
		ulog.Error("Failed to register commands: %v", err)
		return
	}

	mu.RLock()
	count := len(commands)
	mu.RUnlock()

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Successfully registered %d commands!", count))
	ulog.Info("Commands registered via !registerCommands by user %s", m.Author.ID)
}

func handleListCommands(s *discordgo.Session, m *discordgo.MessageCreate) {
	allCommands := GetAllCommands()

	if len(allCommands) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No commands registered.")
		return
	}

	message := fmt.Sprintf("**Registered Commands (%d)**\n```\n", len(allCommands))

	cmdNames := make([]string, 0, len(allCommands))
	for name := range allCommands {
		cmdNames = append(cmdNames, name)
	}

	for i := 0; i < len(cmdNames); i++ {
		for j := i + 1; j < len(cmdNames); j++ {
			if cmdNames[i] > cmdNames[j] {
				cmdNames[i], cmdNames[j] = cmdNames[j], cmdNames[i]
			}
		}
	}

	for _, name := range cmdNames {
		cmd := allCommands[name]
		perms := []string{}
		if cmd.AllowEveryone {
			perms = append(perms, "everyone")
		}
		if cmd.AllowAdmin {
			perms = append(perms, "admin")
		}
		if cmd.AllowStaff {
			perms = append(perms, "staff")
		}

		permStr := ""
		if len(perms) > 0 {
			permStr = fmt.Sprintf(" [%s]", fmt.Sprintf("%v", perms)[1:len(fmt.Sprintf("%v", perms))-1])
		}

		cooldownStr := ""
		if cmd.Cooldown > 0 {
			cooldownStr = fmt.Sprintf(" (cooldown: %ds)", cmd.Cooldown)
		}

		message += fmt.Sprintf("• %s - %s%s%s\n", name, cmd.Description, cooldownStr, permStr)
	}

	message += "```"

	if len(message) > 1900 {
		chunk := "**Registered Commands - Part 1**\n```\n"
		for _, name := range cmdNames {
			cmd := allCommands[name]
			line := fmt.Sprintf("• %s - %s\n", name, cmd.Description)
			if len(chunk)+len(line) > 1900 {
				chunk += "```"
				s.ChannelMessageSend(m.ChannelID, chunk)
				chunk = "```\n"
			}
			chunk += line
		}
		chunk += "```"
		s.ChannelMessageSend(m.ChannelID, chunk)
	} else {
		s.ChannelMessageSend(m.ChannelID, message)
	}

	ulog.Info("Commands listed via !listCommands by user %s", m.Author.ID)
}
