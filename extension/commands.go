package extension

import (
	"fmt"
	"github.com/weaming/telehello/core"
	"strings"
)

type CommandHandler func(body string, params *HandlerParameter) string

type HandlerParameter struct {
	UserID string
	Rss    *RSSPool
	Turing *TuringBot
}

var handlerMap = map[string]CommandHandler{
	"start":    startHandler,
	"admin":    adminHandler,
	"status":   startHandler,
	"storage":  storageHandler,
	"uptime":   uptimeHandler,
	"debug":    debugHandler,
	"weather":  weatherHandler,
	"users":    usersHandler,
	"address":  addrssHandler,
	"forcerss": forcerssHandler,
	"delrss":   delrssHandler,
	"listrss":  listrssHandler,
}

func ProcessCommand(text, userID string, rss *RSSPool, turing *TuringBot) string {
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
			Turing: turing,
		})
	}
	return "未知命令 " + cmd
}

func startHandler(body string, params *HandlerParameter) string {
	return strings.Join([]string{
		"/weather  查天气",
		"/uptime   查看服务器运行总时间",
		"/addrss   添加RSS源",
		"/delrss   删除RSS源",
		"/listrss  列出已添加的源",
		"/forcerss 强制立即抓取",
	}, "\n")
}

func adminHandler(body string, params *HandlerParameter) string {
	return strings.Join([]string{
		"/status   supervisord状态",
		"/storage  查看磁盘空间",
		"/uptime   查看服务器运行总时间",
		"/users    查看从上次运行后新建的客户数",
		"/debug    查看我的ChatID",
	}, "\n")

}

func statusHandler(body string, params *HandlerParameter) string {
	return core.ShellCommand("sudo supervisorctl status")
}

func storageHandler(body string, params *HandlerParameter) string {
	return core.ShellCommand("sudo df -h")
}

func uptimeHandler(body string, params *HandlerParameter) string {
	return core.ShellCommand("uptime")
}

func debugHandler(body string, params *HandlerParameter) string {
	return core.ChatsMap[params.UserID].String()
}

func weatherHandler(body string, params *HandlerParameter) string {
	return params.Turing.Answer("查天气 "+body, params.UserID)
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
