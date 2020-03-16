package extension

import (
	"fmt"
	"strings"

	"github.com/weaming/telehello/core"
)

type CommandHandler func(body string, params *HandlerParameter) string

type HandlerParameter struct {
	UserID string
	Rss    *RSSPool
}

var handlerMap = map[string]CommandHandler{
	"start":    startHandler,
	"debug":    debugHandler,
	"users":    usersHandler,
	"addrss":   addrssHandler,
	"forcerss": forcerssHandler,
	"delrss":   delrssHandler,
	"listrss":  listrssHandler,
}

func ProcessCommand(text, userID string, rss *RSSPool) string {
	var cmd, body string
	split := strings.SplitN(text, " ", 2)
	l := len(split)
	cmd = split[0][1:]

	if l > 1 {
		body = strings.Trim(split[1], " ")
	}

	if h, ok := handlerMap[cmd]; ok {
		return h(body, &HandlerParameter{
			UserID: userID,
			Rss:    rss,
		})
	}
	return "未知命令 " + cmd
}

func startHandler(body string, params *HandlerParameter) string {
	return strings.Join([]string{
		"/addrss   添加RSS源",
		"/delrss   删除RSS源",
		"/listrss  列出已添加的源",
		"/forcerss 强制立即抓取",
	}, "\n")
}

func debugHandler(body string, params *HandlerParameter) string {
	return core.ChatsMap[params.UserID].String()
}

func usersHandler(body string, params *HandlerParameter) string {
	if admin, ok := core.ChatsMap[core.AdminKey]; ok {
		if params.UserID == admin.ID {
			chats := []string{"New users since last starting running:"}
			for _, chat := range core.ChatsMap {
				chats = append(chats, fmt.Sprintf("%v(%v)", chat.TeleName, chat.ID))
			}
			chats = append(chats, "Chats IDs in DB:")
			chats = core.ExtendStringList(chats, params.Rss.GetChatIDList())
			return strings.Join(chats, "\n")
		}
	}
	return "Only administrator can view users"
}

func addrssHandler(body string, params *HandlerParameter) string {
	if body != "" {
		urls := strings.Fields(body)
		for _, u := range urls {
			err := params.Rss.AddRSS(params.UserID, u)
			core.NotifiedErr(err, params.UserID)
		}
		return "received rss:\n" + strings.Join(urls, "\n")
	} else {
		return "未给出RSS URL"
	}
}

func forcerssHandler(body string, params *HandlerParameter) string {
	params.Rss.CrawlForUser(params.UserID, false)
	return "force crawled your RSSes"
}

func delrssHandler(body string, params *HandlerParameter) string {
	if body != "" {
		urls := strings.Fields(body)
		for _, u := range urls {
			err := params.Rss.DeleteRSS(params.UserID, u)
			core.NotifiedErr(err, params.UserID)
		}
		return "deleted rss:\n" + strings.Join(urls, "\n")
	} else {
		return "未给出RSS URL"
	}
}

func listrssHandler(body string, params *HandlerParameter) string {
	urls, err := params.Rss.GetOldURLs(params.UserID)
	if err != nil {
		return "error getting RSS: " + err.Error()
	}
	if len(urls) == 0 {
		return "haven't received any additional RSS"
	}
	return "received rss:\n" + strings.Join(urls, "\n")
}
