package stats

import (
	"sort"
	"strings"
)

var (
	tagsCache = make(map[string]map[string]string)
)

type Tags struct {
	tags map[string]string
	text string
}

func NewTags(tags ...string) *Tags {
	t := &Tags{
		tags: make(map[string]string),
	}
	for i := 0; i+1 < len(tags); i += 2 {
		t.tags[tags[i]] = tags[i+1]
	}
	return t
}

func (tags *Tags) Add(name, value string) {
	tags.tags[name] = value
}

func (tags *Tags) String() string {
	if tags.text == "" {
		ss := sort.StringSlice(make([]string, len(tags.tags)))
		i := 0
		for key, value := range tags.tags {
			ss[i] = key + "=" + value
		}
		ss.Sort()
		tags.text = strings.Join(ss, ",")
		tagsCache[tags.text] = tags.tags
	}
	return tags.text
}
