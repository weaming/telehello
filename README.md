## How to install

    go get -u github.com/weaming/telehello

## Usage

```
一个Telegram消息机器人

Features:
	1. RSS抓取
	2. HTTP接口接收消息
	3. 图灵聊天机器人
	4. 执行服务器脚本

Usage of telehello:
  -douban float
    	douban movie min score (default 8)
  -l string
    	[host]:port http hook to receive message (default ":1234")
  -rss int
    	period of crawling wanqu.co RSS in minutes (default 1440)
  -t int
    	telegram bot long poll timeout in seconds (default 30)
  -telegramID string
    	your telegram ID without @ (default "weaming")
  -x	delete bot status KV database before start
```
