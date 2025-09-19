package utils

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	Logger *log.Logger
	once   sync.Once
	loc    = time.FixedZone("CST", 8*3600)
)

func InitLog(logFile string) error {
	var err error
	once.Do(func() {
		// 确保日志目录存在
		dir := filepath.Dir(logFile)
		if dir != "." && dir != "/" {
			if e := os.MkdirAll(dir, 0755); e != nil {
				err = e
				return
			}
		}
		
		f, e := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if e != nil {
			err = e
			return
		}
		mw := io.MultiWriter(f, os.Stdout)
		Logger = log.New(mw, "", 0)
	})
	return err
}

func Info(v ...interface{}) {
	if Logger != nil {
		now := time.Now().In(loc).Format("2006-01-02 15:04:05")
		Logger.Println(append([]interface{}{now}, v...)...)
	}
}

func Error(v ...interface{}) {
	if Logger != nil {
		now := time.Now().In(loc).Format("2006-01-02 15:04:05")
		args := make([]interface{}, 0, 2+len(v))
		args = append(args, now)
		args = append(args, "[ERROR]")
		args = append(args, v...)
		Logger.Println(args...)
	}
}

func Infof(format string, v ...interface{}) {
	if Logger != nil {
		now := time.Now().In(loc).Format("2006-01-02 15:04:05")
		Logger.Printf("["+now+"] [INFO] "+format, v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if Logger != nil {
		now := time.Now().In(loc).Format("2006-01-02 15:04:05")
		Logger.Printf("["+now+"] [ERROR] "+format, v...)
	}
}
