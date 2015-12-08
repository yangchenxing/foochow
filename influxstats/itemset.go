package stats

type ItemSet map[string]map[string]map[string]interface{}

func (set ItemSet) getFields(measurement string, tags *Tags) map[string]interface{} {
	tagItems := set[measurement]
	if tagItems == nil {
		tagItems = make(map[string]map[string]interface{})
		set[measurement] = tagItems
	}
	fields := tagItems[tags.String()]
	if fields == nil {
		fields = make(map[string]interface{})
		tagItems[tags.String()] = fields
	}
	return fields
}

func (set ItemSet) AddFloat(measurement string, tags *Tags, name string, value float64) {
	fields := set.getFields(measurement, tags)
	if v, ok := fields[name].(float64); ok {
		fields[name] = v + value
	} else {
		fields[name] = value
	}
}

func (set ItemSet) AddInt(measurement string, tags *Tags, name string, value int64) {
	fields := set.getFields(measurement, tags)
	if v, ok := fields[name].(int64); ok {
		fields[name] = v + value
	} else {
		fields[name] = value
	}
}

func (set ItemSet) SetFloat(measurement string, tags *Tags, name string, value float64) {
	set.getFields(measurement, tags)[name] = value
}

func (set ItemSet) SetInt(measurement string, tags *Tags, name string, value int64) {
	set.getFields(measurement, tags)[name] = value
}

func (set ItemSet) GetFloat(measurement string, tags *Tags, name string) float64 {
	if v := set.getFields(measurement, tags)[name]; v == nil {
		return 0
	} else {
		return v.(float64)
	}
}

func (set ItemSet) GetInt(measurement string, tags *Tags, name string) int64 {
	if v := set.getFields(measurement, tags)[name]; v == nil {
		return 0
	} else {
		return v.(int64)
	}
}
