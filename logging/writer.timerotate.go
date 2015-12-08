package logging

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type TimeRotateWriter struct {
	sync.Mutex
	Path     string
	Split    string
	Interval time.Duration
	Keep     time.Duration
	file     *os.File
}

func (writer *TimeRotateWriter) initialize() error {
	var err error
	// 若文件已经存在，则判断是否需要重命名为一个切分文件
	if info, err := os.Stat(writer.Path); err == nil {
		timestamp := info.ModTime().Truncate(writer.Interval)
		if timestamp.Before(time.Now().Truncate(writer.Interval)) {
			os.Rename(writer.Path, timestamp.Format(writer.Split))
		}
	}
	// 打开文件
	if writer.file, err = os.OpenFile(writer.Path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755); err != nil {
		return err
	}
	// 启动定时切割和清理
	go func() {
		timestamp := time.Now().Truncate(writer.Interval)
		for {
			time.Sleep(timestamp.Add(writer.Interval).Sub(time.Now()))
			writer.rotate(timestamp)
			writer.clean(timestamp)
			timestamp = timestamp.Add(writer.Interval)
		}
	}()
	return nil
}

func (writer *TimeRotateWriter) rotate(timestamp time.Time) {
	writer.Lock()
	defer writer.Unlock()

	splitPath := timestamp.Format(writer.Split)
	if _, err := os.Stat(splitPath); err == nil {
		fmt.Fprintf(os.Stderr, "日志切割出错，切割文件已经存在: split=%q\n", splitPath)
		return
	}
	writer.file.Close()
	if err := os.Rename(writer.Path, splitPath); err != nil {
		fmt.Fprintf(os.Stderr, "日志切割出错，重命名失败: path=%q, split=%q, error=%q\n",
			writer.Path, splitPath, err.Error())
		writer.file, _ = os.OpenFile(writer.Path, os.O_WRONLY|os.O_APPEND, 0755)
		return
	}
	var err error
	if writer.file, err = os.OpenFile(writer.Path, os.O_WRONLY|os.O_CREATE, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "日志切割出错，打开文件失败: path=%q, error=%q\n",
			writer.Path, err.Error())
		return
	}
	fmt.Fprintf(os.Stderr, "日志切割成功: path=%q, split=%q\n", writer.Path, splitPath)
}

func (writer *TimeRotateWriter) clean(timestamp time.Time) {
	// 读取切割文件目录
	splitDir := filepath.Dir(writer.Split)
	infos, err := ioutil.ReadDir(splitDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "清理日志出错，无法读取切割文件目录: error=%q\n", err.Error())
		return
	}
	// 逐条匹配时间
	pattern := filepath.Base(writer.Split)
	for _, info := range infos {
		splitTime, err := time.Parse(pattern, info.Name())
		if err != nil {
			continue
		} else if timestamp.Sub(splitTime) <= writer.Keep {
			continue
		}
		splitPath := filepath.Join(splitDir, info.Name())
		if err := os.Remove(splitPath); err != nil {
			fmt.Fprintf(os.Stderr, "清理日志出错，删除失败: path=%q, error=%q\n", splitPath, err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "清理日志成功: path=%q\n", splitPath)
		}
	}
}

func (writer *TimeRotateWriter) Write(bytes []byte) (int, error) {
	writer.Lock()
	defer writer.Unlock()
	defer writer.file.Sync()
	return writer.file.Write(bytes)
}
