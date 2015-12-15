package useragent

import "strings"

type browserSpec struct {
	name     string
	keywords []string
}

var (
	browserSpecs = []browserSpec{
		browserSpec{
			name:     "Baidu",
			keywords: []string{"baidu", "Baidu"}, // 这里存疑
		},
		browserSpec{
			name:     "360",
			keywords: []string{"360browser", "360 Aphone Browser"},
		},
		browserSpec{
			name:     "MQQBrowser",
			keywords: []string{"MQQBrowser"},
		},
		browserSpec{
			name:     "QQLiveBrowser",
			keywords: []string{"QQLiveBrowser", "QQLiveHDBrowser"},
		},
		browserSpec{
			name:     "QQBrowser",
			keywords: []string{"QQBrowser"},
		},
		browserSpec{
			name:     "QQ",
			keywords: []string{" QQ/"},
		},
		browserSpec{
			name:     "MicroMessenger",
			keywords: []string{"MicroMessenger"},
		},
		browserSpec{
			name:     "UCBrowser",
			keywords: []string{"UCBrowser", "UCWEB", "UC ", "UC/"},
		},
		browserSpec{
			name:     "MQQBrowser",
			keywords: []string{"MQQBrowser"},
		},
		browserSpec{
			name:     "Opera",
			keywords: []string{"Opera ", "Opera/"},
		},
		browserSpec{
			name:     "Sogou",
			keywords: []string{"Sogou"},
		},
		browserSpec{
			name:     "Liebao",
			keywords: []string{"LieBao", "liebao"},
		},
		browserSpec{
			name:     "Chrome",
			keywords: []string{"Chrome"},
		},
		browserSpec{
			name:     "Safari",
			keywords: []string{"Safari"},
		},
	}
)

func (ua UserAgent) Browser() string {
	for _, spec := range browserSpecs {
		for _, keyword := range spec.keywords {
			if strings.Contains(string(ua), keyword) {
				return spec.name
			}
		}
	}
	return ""
}
