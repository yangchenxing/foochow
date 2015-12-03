package filewatch

import (
	"fmt"
	"os"
	"time"
)

func WatchFileUpdate(path string, interval time.Duration, callback func(path string)) error {
	if interval < time.Second {
		return fmt.Errorf("检查间隔不能低于1秒")
	}
	var lastModTime time.Time
	if info, err := os.Lstat(path); err != nil {
		return fmt.Errorf("检查文件信息出错: path=%q, error=%q", path, err.Error())
	} else {
		lastModTime = info.ModTime()
	}
	go func() {
		for {
			time.Sleep(interval)
			if info, err := os.Lstat(path); err == nil {
				if lastModTime.Before(info.ModTime()) {
					callback(path)
					lastModTime = info.ModTime()
				}
			}
		}
	}()
	return nil
}
