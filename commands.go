package main

import (
	"fmt"
	"strings"
	"time"
)

func processCommand(text, userID, userName string) string {
	var cmd, body string
	split := strings.SplitN(text, " ", 2)
	l := len(split)
	cmd = split[0][1:]

	if l > 1 {
		body = strings.Trim(split[1], " ")
	}

	if cmd == "start" {
		return strings.Join([]string{
			"/weather  查天气",
			"/uptime   查看服务器运行总时间",
			"/addrss   添加RSS源",
			"/delrss   删除RSS源",
			"/listrss  列出已添加的源",
		}, "\n")
	} else if cmd == "admin" {
		return strings.Join([]string{
			"/status   supervisord状态",
			"/storage  查看磁盘空间",
			"/uptime   查看服务器运行总时间",
			"/users 查看从上次运行后新建的客户数",
			"/debug    查看我的ChatID",
		}, "\n")

	} else if cmd == "status" {
		return ShellCommand("sudo supervisorctl status")
	} else if cmd == "storage" {
		return ShellCommand("sudo df -h")
	} else if cmd == "uptime" {
		return ShellCommand("uptime")

	} else if cmd == "debug" {
		return ChatsMap[userID].String()
	} else if cmd == "weather" {
		return turing.answer("查天气 "+body, userID)
	} else if cmd == "users" {
		if admin, ok := ChatsMap[AdminKey]; ok {
			if userID == admin.ID {
				var chats []string
				for _, chat := range ChatsMap {
					chats = append(chats, fmt.Sprintf("%v(%v)", chat.TeleName, chat.ID))
				}
				return strings.Join(chats, "\n")
			}
		}
		return "Only administrator can view users"
	} else if cmd == "addrss" {
		if l == 2 {
			urls := strings.Fields(body)
			for _, u := range urls {
				err := AddRSS(userID, u, time.Minute*time.Duration(scanMinutes))
				NotifyErr(err, userID)
			}
			return "received rss:\n" + strings.Join(urls, "\n")
		} else {
			return "未给出RSS URL"
		}
	} else if cmd == "forcerss" {
		CrawlerForUser(userID, false)
		NotifyText("force crawled your RSSes", userID)
	} else if cmd == "delrss" {
		if l == 2 {
			urls := strings.Fields(body)
			for _, u := range urls {
				err := DeleteRSS(userID, u)
				NotifyErr(err, userID)
			}
			return "received rss:\n" + strings.Join(urls, "\n")
		} else {
			return "未给出RSS URL"
		}
	} else if cmd == "listrss" {
		urls, err := GetOldURLs(userID)
		if err != nil {
			return "error getting RSS: " + err.Error()
		}
		if len(urls) == 0 {
			return "haven't receive any additional RSS"
		}
		return "received rss:\n" + strings.Join(urls, "\n")
	}
	return "未知命令 " + cmd
}
