package logging

import (
	"bytes"
	"container/list"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type EMailWriter struct {
	Server    string
	Sender    string
	Password  string
	Receivers []string
	Subject   string
	Delay     time.Duration
	ch        chan []byte
}

func (writer *EMailWriter) initialize() {
	// 启动发送守候
	writer.ch = make(chan []byte)
	go func() {
		messages := list.New()
		for {
			messages.Init()
			messages.PushBack(<-writer.ch)
			ticker := time.NewTicker(writer.Delay)
			for delay := true; delay; {
				select {
				case message := <-writer.ch:
					messages.PushBack(message)
				case <-ticker.C:
					ticker.Stop()
					delay = false
				}
			}
			var buf bytes.Buffer
			fmt.Fprintf(&buf, "From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n",
				writer.Sender,
				strings.Join(writer.Receivers, ","),
				writer.Subject)
			for elem := messages.Front(); elem != nil; elem = elem.Next() {
				buf.Write(elem.Value.([]byte))
			}
			auth := smtp.PlainAuth(writer.Sender, writer.Sender, writer.Password,
				strings.Split(writer.Server, ":")[0])
			if err := smtp.SendMail(writer.Server, auth, writer.Sender,
				writer.Receivers, buf.Bytes()); err != nil {
				fmt.Fprintf(os.Stderr, "发送日志邮件出错: subject=%q, error=%q\n",
					writer.Subject, err.Error())
			} else {
				fmt.Fprintf(os.Stderr, "发送日志邮件成功: subject=%q\n", writer.Subject)
			}
		}
	}()
}

func (writer *EMailWriter) Write(bytes []byte) (int, error) {
	writer.ch <- bytes
	return len(bytes), nil
}
