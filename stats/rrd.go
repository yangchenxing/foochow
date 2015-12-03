package stats

import (
	"container/list"
	"sync"
	"time"
)

const (
	defaultSerialSize = 3601
)

var (
	items      = make(map[string]StatsItem)
	data       = list.New()
	current    int
	rrdLock    sync.Mutex
	serialSize = defaultSerialSize
)

func init() {
	go func() {
		for {
			time.Sleep(time.Second)
			next := (current + 1) % serialSize
			for elem := data.Front(); elem != nil; elem = elem.Next() {
				(elem.Value.([]float64))[next] = 0
			}
			rrdLock.Lock()
			current = next
			rrdLock.Unlock()
		}
	}()
}

func newData() []float64 {
	values := make([]float64, serialSize)
	data.PushBack(values)
	return values
}

func registerItem(name string, item StatsItem) {
	if name != "" {
		items[name] = item
	}
}

func LockRRD(callback func()) {
	rrdLock.Lock()
	defer rrdLock.Unlock()
	callback()
}

func GetStatsItem(name string) StatsItem {
	return items[name]
}

func GetAllStatsItemValues(duration time.Duration) map[string]float64 {
	rrdLock.Lock()
	defer rrdLock.Unlock()
	result := make(map[string]float64)
	for name, item := range items {
		result[name] = item.Get(duration)
	}
	return result
}
