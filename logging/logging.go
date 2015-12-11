package logging

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	gopath      string
	ip          string
	hostname, _ = os.Hostname()
	handlers    = make(map[string][]*Handler)
)

type Logger func(level, pattern string, args ...interface{})

func init() {
	// 获取GOPATH
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	gopath = file[:len(dir)-len("/github.com/yangchenxing/foochow/logging")]
	// 获取本地IP
	ip = func() string {
		if infs, err := net.Interfaces(); err == nil && len(infs) > 0 {
			for _, inf := range infs {
				if inf.Flags&net.FlagLoopback != 0 {
					continue
				}
				if addrs, err := inf.Addrs(); err == nil && len(addrs) > 0 {
					for _, addr := range addrs {
						if ipnet, ok := addr.(*net.IPNet); ok {
							if ipv4 := ipnet.IP.To4(); ipv4 != nil {
								return ipv4.String()
							}
						}
					}
				}
			}
		}
		return ""
	}()
	// 构造默认handler
	defaultHandler := &Handler{
		Levels:  []string{"debug", "info", "warn", "error", "fatal"},
		Format:  "$level [$time][$file:$line][$func] $message",
		Writers: []Writer{os.Stderr},
	}
	defaultHandler.initialize()
	AddHandler(defaultHandler)
}

func AddHandler(handler *Handler) {
	if handler == nil {
		return
	}
	for _, level := range handler.Levels {
		if handlers[level] == nil {
			handlers[level] = make([]*Handler, 0, 4)
		}
		handlers[level] = append(handlers[level], handler)
	}
}

func LogEx(skip int, level string, custom map[string]string, format string, args ...interface{}) {
	targetHandlers := handlers[level]
	wildHandlers := handlers["*"]
	if len(targetHandlers) == 0 && len(wildHandlers) == 0 {
		return
	}
	// 采集数据
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = ""
		line = 0
	}
	funcname := ""
	if caller := runtime.FuncForPC(pc); caller != nil {
		funcname = caller.Name()
	}
	if strings.HasPrefix(file, gopath) {
		file = file[len(gopath)+1:]
	}
	event := map[string]string{
		"level":   level,
		"message": fmt.Sprintf(format, args...),
		"file":    file,
		"line":    strconv.Itoa(line),
		"time":    time.Now().Format("2006-01-02:15:04:05-0700"),
		"func":    funcname,
		"host":    hostname,
		"ip":      ip,
	}
	for key, value := range custom {
		event[key] = value
	}
	for _, handler := range targetHandlers {
		handler.write(event)
	}
	for _, handler := range wildHandlers {
		handler.write(event)
	}
}

func Log(level, format string, args ...interface{}) {
	LogEx(2, level, nil, format, args...)
}

func Debug(format string, args ...interface{}) {
	LogEx(2, "debug", nil, format, args...)
}

func Info(format string, args ...interface{}) {
	LogEx(2, "info", nil, format, args...)
}

func Warn(format string, args ...interface{}) {
	LogEx(2, "warn", nil, format, args...)
}

func Error(format string, args ...interface{}) {
	LogEx(2, "error", nil, format, args...)
}

func Fatal(format string, args ...interface{}) {
	LogEx(2, "fatal", nil, format, args...)
}
