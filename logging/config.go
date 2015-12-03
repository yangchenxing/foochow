package logging

type Config struct {
	Handlers []*Handler
}

func LoadConfig(config *Config) {
	newHandlers := make(map[string][]*Handler)
	for _, handler := range config.Handlers {
		handler.initialize()
		for _, level := range handler.Levels {
			if newHandlers[level] == nil {
				newHandlers[level] = make([]*Handler, 0, 4)
			}
			newHandlers[level] = append(newHandlers[level], handler)
		}
	}
	handlers = newHandlers
}
