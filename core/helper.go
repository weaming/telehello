package core

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/tucnak/telebot"
)

func FatalErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func PrintErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

func LogMessage(msg telebot.Message) {
	log.Printf(">> from %v: %v", prettyUser(msg.Origin()), msg.Text)
	//log.Printf("%+v", msg)
}

func prettyUser(user telebot.User) string {
	return fmt.Sprintf("<%d %v>(%v %v)", user.ID, user.Username, user.FirstName, user.LastName)
}

func prettyTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
	return fmt.Sprintf("%v-%v-%v %v:%v:%v",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

func DownloadTeleFile(url, fileName string) (filePath string, err error) {
	// parse name
	split := strings.SplitN(fileName, "_", 2)
	if len(split) == 2 {
		filePath = path.Join("/tmp", "telebot", split[0], split[1])
	} else {
		filePath = path.Join("/tmp", "telebot", fileName)
	}

	return DownloadURL(url, filePath)
}

func DownloadURL(url, filePath string) (string, error) {
	// if not absolute
	if !strings.HasPrefix(filePath, "/") {
		if strings.HasPrefix(filePath, "~/") {
			usr, _ := user.Current()
			home := usr.HomeDir
			filePath = filepath.Join(home, filePath[2:])
		} else {
			return "", errors.New("file path must starts with '/' or '~/'")
		}
	}
	//fmt.Println(url, filePath)

	// if exist, return error
	if exist, _ := ExistFile(filePath); exist {
		return filePath, errors.New(fmt.Sprintf("existed '%v'", filePath))
	}
	// prepare directory
	if exist, _ := ExistFile(path.Dir(filePath)); !exist {
		err := os.MkdirAll(path.Dir(filePath), 0777)
		if err != nil {
			return "", err
		}
	}

	// http get
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// write to file
	err = ioutil.WriteFile(filePath, bytes, 0666)
	if err != nil {
		return "", err
	} else {
		return filePath, err
	}
}

func ExistFile(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func MaxFileSize(files []telebot.Thumbnail) (f telebot.Thumbnail, err error) {
	l := len(files)
	if l == 0 {
		return f, errors.New("no files")
	}

	f = files[0]
	for i := 1; i < l; i++ {
		if files[i].FileSize > f.FileSize {
			f = files[i]
		}
	}
	return
}

func ExtendStringList(a, b []string) []string {
	for _, x := range b {
		a = append(a, x)
	}

	return a
}

func GetMessageIP(req *http.Request) string {
	realIP := req.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	return strings.Split(req.RemoteAddr, ":")[0]
}
