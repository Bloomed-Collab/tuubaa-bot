package ticket

import (
	"bytes"
	"fmt"
	"html"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
)

func fetchMessages(s *discordgo.Session, channelID string) ([]*discordgo.Message, error) {
	var all []*discordgo.Message
	beforeID := ""
	for {
		var msgs []*discordgo.Message
		var err error
		if beforeID == "" {
			msgs, err = s.ChannelMessages(channelID, 100, "", "", "")
		} else {
			msgs, err = s.ChannelMessages(channelID, 100, beforeID, "", "")
		}
		if err != nil {
			return all, err
		}
		if len(msgs) == 0 {
			break
		}
		all = append(all, msgs...)
		beforeID = msgs[len(msgs)-1].ID
		if len(msgs) < 100 {
			break
		}
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].Timestamp.Before(all[j].Timestamp)
	})
	return all, nil
}

func buildTXT(msgs []*discordgo.Message) []byte {
	var buf bytes.Buffer
	for _, m := range msgs {
		name := "Unknown"
		if m.Author != nil {
			name = m.Author.Username
		}
		ts := m.Timestamp.In(berlinLoc).Format("02.01.2006 15:04")
		buf.WriteString(fmt.Sprintf("[%s] %s: %s\n", ts, name, m.Content))
		for _, a := range m.Attachments {
			buf.WriteString(fmt.Sprintf("  [Attachment: %s]\n", a.URL))
		}
	}
	return buf.Bytes()
}

func buildHTML(msgs []*discordgo.Message, ticketKind string, openedByName string, openedAt time.Time, closedByName string, closedAt time.Time) []byte {
	var buf bytes.Buffer
	buf.WriteString(`<!DOCTYPE html>
<html lang="de">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Ticket Transcript</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    background: #313338;
    color: #dbdee1;
    font-family: 'Segoe UI', 'Helvetica Neue', Helvetica, Arial, sans-serif;
    font-size: 15px;
    line-height: 1.375;
    padding: 0;
  }
  .header {
    background: #2b2d31;
    padding: 20px 24px;
    border-bottom: 1px solid #1e1f22;
  }
  .header h1 {
    font-size: 20px;
    color: #f2f3f5;
    margin-bottom: 8px;
  }
  .header .meta {
    font-size: 13px;
    color: #949ba4;
    line-height: 1.6;
  }
  .header .meta span {
    color: #dbdee1;
    font-weight: 500;
  }
  .messages {
    padding: 16px 0;
  }
  .msg {
    padding: 2px 24px;
    display: flex;
    gap: 16px;
    margin-top: 4px;
  }
  .msg:hover {
    background: #2e3035;
  }
  .msg.first-in-group {
    margin-top: 17px;
  }
  .avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    flex-shrink: 0;
    background: #5865f2;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-weight: 600;
    font-size: 16px;
  }
  .avatar img {
    width: 40px;
    height: 40px;
    border-radius: 50%;
  }
  .content {
    min-width: 0;
    flex: 1;
  }
  .author-row {
    display: flex;
    align-items: baseline;
    gap: 8px;
  }
  .author {
    font-weight: 600;
    color: #f2f3f5;
    font-size: 15px;
  }
  .author.bot {
    color: #5865f2;
  }
  .timestamp {
    font-size: 11px;
    color: #949ba4;
  }
  .text {
    color: #dbdee1;
    word-wrap: break-word;
    white-space: pre-wrap;
  }
  .attachment {
    margin-top: 4px;
    padding: 8px 12px;
    background: #2b2d31;
    border: 1px solid #1e1f22;
    border-radius: 8px;
    display: inline-block;
  }
  .attachment a {
    color: #00a8fc;
    text-decoration: none;
  }
  .attachment a:hover {
    text-decoration: underline;
  }
  .footer {
    background: #2b2d31;
    padding: 16px 24px;
    border-top: 1px solid #1e1f22;
    font-size: 13px;
    color: #949ba4;
    text-align: center;
  }
</style>
</head>
<body>
`)

	k := ticketKinds[ticketKind]
	buf.WriteString(fmt.Sprintf(`<div class="header">
<h1>%s %s — Ticket Transcript</h1>
<div class="meta">
  <b>Geöffnet von:</b> <span>%s</span> — %s<br>
  <b>Geschlossen von:</b> <span>%s</span> — %s<br>
  <b>Nachrichten:</b> <span>%d</span>
</div>
</div>
<div class="messages">
`,
		html.EscapeString(k.icon),
		html.EscapeString(k.title),
		html.EscapeString(openedByName),
		openedAt.In(berlinLoc).Format("02.01.2006 15:04"),
		html.EscapeString(closedByName),
		closedAt.In(berlinLoc).Format("02.01.2006 15:04"),
		len(msgs),
	))

	var lastAuthorID string
	for _, m := range msgs {
		authorName := "Unknown"
		authorID := ""
		isBot := false
		avatarHTML := `<div class="avatar">?</div>`
		if m.Author != nil {
			authorName = m.Author.Username
			authorID = m.Author.ID
			isBot = m.Author.Bot
			if m.Author.AvatarURL("64") != "" {
				avatarHTML = fmt.Sprintf(`<div class="avatar"><img src="%s" alt="%s"></div>`,
					html.EscapeString(m.Author.AvatarURL("64")),
					html.EscapeString(authorName))
			} else {
				initial := "?"
				if len(authorName) > 0 {
					initial = string([]rune(authorName)[:1])
				}
				avatarHTML = fmt.Sprintf(`<div class="avatar">%s</div>`, html.EscapeString(initial))
			}
		}

		isNewGroup := authorID != lastAuthorID
		lastAuthorID = authorID

		groupClass := ""
		if isNewGroup {
			groupClass = " first-in-group"
		}

		ts := m.Timestamp.In(berlinLoc).Format("02.01.2006 15:04")

		buf.WriteString(fmt.Sprintf(`<div class="msg%s">`, groupClass))

		if isNewGroup {
			buf.WriteString(avatarHTML)
			buf.WriteString(`<div class="content">`)
			botClass := ""
			if isBot {
				botClass = " bot"
			}
			buf.WriteString(fmt.Sprintf(`<div class="author-row"><span class="author%s">%s</span><span class="timestamp">%s</span></div>`,
				botClass,
				html.EscapeString(authorName),
				ts))
		} else {
			buf.WriteString(`<div style="width:40px;flex-shrink:0"></div><div class="content">`)
		}

		if m.Content != "" {
			buf.WriteString(fmt.Sprintf(`<div class="text">%s</div>`, html.EscapeString(m.Content)))
		}

		for _, a := range m.Attachments {
			buf.WriteString(fmt.Sprintf(`<div class="attachment"><a href="%s" target="_blank">📎 %s</a></div>`,
				html.EscapeString(a.URL),
				html.EscapeString(a.Filename)))
		}

		buf.WriteString(`</div></div>` + "\n")
	}

	buf.WriteString(`</div>
<div class="footer">Ticket Transcript — tuubaa Bot</div>
</body>
</html>`)

	return buf.Bytes()
}
