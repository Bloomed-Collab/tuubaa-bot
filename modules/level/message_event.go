package level

import (
	"sync"
	"time"

	"github.com/S42yt/tuubaa-bot/core"
	"github.com/bwmarrin/discordgo"
)

var (
	srvTS = newDeque[time.Time]()
	srvMu sync.Mutex

	usrTS = map[string]*deque[time.Time]{}
	usrMu sync.Mutex

	anHour = int64(60)
	aDay   = int64(1440)
	aWeek  = int64(10080)
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
	if last, ok := usrTS[userID].peekAt(0); ok && now.Sub(last) < 10*time.Second {
		usrMu.Unlock()
		return
	}
	usrTS[userID].pushFront(now)
	usrMu.Unlock()

	srvMu.Lock()
	srvTS.pushFront(now)
	srvMu.Unlock()

	srvMu.Lock()
	for {
		last, ok := srvTS.peekBack()
		if !ok || last.UnixMilli() >= now.UnixMilli()-anHour {
			break
		}
		_, _ = srvTS.popBack()
	}
	serverHour := 0
	for idx := 0; idx < srvTS.size(); idx++ {
		if ts, ok := srvTS.peekAt(idx); ok && ts.UnixMilli() > now.UnixMilli()-anHour {
			serverHour = idx
			break
		}
	}
	srvMu.Unlock()

	usrMu.Lock()
	for _, q := range usrTS {
		for {
			last, ok := q.peekBack()
			if !ok || last.UnixMilli() >= now.UnixMilli()-aDay {
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
				continue
			}
			if userDaily == 0 && ts.UnixMilli() > now.UnixMilli()-aDay {
				userDaily = idx
			}
			if ts.UnixMilli() > now.UnixMilli()-aWeek {
				userWeekly = idx
				break
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
