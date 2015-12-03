package logging

import (
	"bytes"
	"io"
	"regexp"
)

var (
	holderPattern = regexp.MustCompile("\\$[a-zA-Z_][a-zA-Z0-9_]*")
)

type Writer io.Writer

type Handler struct {
	Levels  []string
	Format  string
	Writers []Writer
	pattern []string
}

func (handler *Handler) initialize() {
	submatches := holderPattern.FindAllString(handler.Format, -1)
	texts := holderPattern.Split(handler.Format, -1)
	handler.pattern = make([]string, len(submatches)+len(texts))
	for i := 0; i < len(handler.pattern); i++ {
		if i%2 == 0 {
			handler.pattern[i] = texts[i/2]
		} else {
			handler.pattern[i] = submatches[i/2][1:]
		}
	}
}

func (handler *Handler) write(event map[string]string) error {
	var buf bytes.Buffer
	for i, text := range handler.pattern {
		if i%2 == 0 {
			buf.WriteString(text)
		} else {
			buf.WriteString(event[text])
		}
	}
	buf.WriteRune('\n')
	text := buf.Bytes()
	for _, writer := range handler.Writers {
		if _, err := writer.Write(text); err != nil {
			return err
		}
	}
	return nil
}
