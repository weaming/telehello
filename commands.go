package main

import (
	"strings"
	"time"
)

func processCommand(text, id string) string {
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
			"/turing   和机器人聊天",
			"/update   更新TeleBot代码并重启",
			"/status   supervisord状态",
			"/storage  查看磁盘空间",
			"/uptime   查看服务器运行总时间",
			"/addrss   添加RSS源",
			"/listrss  列出已添加的源",
		}, "\n")
	} else if cmd == "update" {
		return ShellScript("/etc/updateTelegramBot.sh")
	} else if cmd == "status" {
		return ShellCommand("sudo supervisorctl status")
	} else if cmd == "storage" {
		return ShellCommand("sudo df -h")
	} else if cmd == "uptime" {
		return ShellCommand("uptime")
	} else if cmd == "turing" {
		if l == 2 {
			return turing.answer(body, id)
		} else {
			return "要跟图灵机器人聊天，可直接回复你想说的话"
		}
	} else if cmd == "weather" {
		return turing.answer("查天气 "+body, id)
	} else if cmd == "shell" {
		if l == 2 {
			return ShellCommand(body)
		} else {
			return "未给出命令"
		}
	} else if cmd == "script" {
		if l == 2 {
			return ShellScript(body)
		} else {
			return "请给出shell脚本完整路径"
		}
	} else if cmd == "download" {
		if l == 2 {
			urlTarget := strings.Fields(body)
			if len(urlTarget) < 2 {
				return "请给出URL和保存绝对路径，空格分隔"
			}
			outPath, err := downloadURL(urlTarget[0], urlTarget[1])
			if err != nil {
				return "error: " + err.Error()
			}
			return "downloaded at " + outPath
		} else {
			return "未给出下载URL"

		}
	} else if cmd == "addrss" {
		if l == 2 {
			urls := strings.Fields(body)
			for _, u := range urls {
				err := AddRSS(u, time.Minute*time.Duration(scanMinutes))
				NotifyErr(err)
			}
			return "received rss:\n" + strings.Join(urls, "\n")
		} else {
			return "未给出RSS URL"

		}
	} else if cmd == "listrss" {
		urls, err := GetOldURLs()
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
