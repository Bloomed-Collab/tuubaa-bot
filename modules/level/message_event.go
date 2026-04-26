package level

import (
	"sync"
	"time"

	"github.com/S42yt/tuubaa-bot/core"
	"github.com/bwmarrin/discordgo"
)

const xpCooldown = 60 * time.Second

var (
	srvTS = newDeque[time.Time]()
	srvMu sync.Mutex

	usrTS = map[string]*deque[time.Time]{}
	usrMu sync.Mutex
)

func init() {
	core.On(messageXPHandler)
}

func messageXPHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author == nil || m.Author.Bot || m.Member == nil || m.GuildID == "" {
		return
	}

	now := time.Now()
	userID := m.Author.ID

	usrMu.Lock()
	if usrTS[userID] == nil {
		usrTS[userID] = newDeque[time.Time]()
	}
	if last, ok := usrTS[userID].peekAt(0); ok && now.Sub(last) < xpCooldown {
		usrMu.Unlock()
		return
	}
	usrTS[userID].pushFront(now)
	usrMu.Unlock()

	srvMu.Lock()
	srvTS.pushFront(now)
	for {
		last, ok := srvTS.peekBack()
		if !ok || now.Sub(last) <= time.Hour {
			break
		}
		_, _ = srvTS.popBack()
	}
	serverHour := srvTS.size()
	srvMu.Unlock()

	usrMu.Lock()
	for _, q := range usrTS {
		for {
			last, ok := q.peekBack()
			if !ok || now.Sub(last) <= 7*24*time.Hour {
				break
			}
			_, _ = q.popBack()
		}
	}
	userDaily := 0
	userWeekly := 0
	if q := usrTS[userID]; q != nil {
		for idx := 0; idx < q.size(); idx++ {
			ts, ok := q.peekAt(idx)
			if !ok {
				break
			}
			age := now.Sub(ts)
			if age > 7*24*time.Hour {
				break
			}
			userWeekly++
			if age <= 24*time.Hour {
				userDaily++
			}
		}
	}
	usrMu.Unlock()

	h, min_, _ := now.Clock()
	daytime := h*60 + min_

	xpGained := evalTextXP(len(m.Content), daytime, serverHour, userDaily, userWeekly)
	guildID := m.GuildID

	go addXP(s, guildID, userID, xpGained)
}
