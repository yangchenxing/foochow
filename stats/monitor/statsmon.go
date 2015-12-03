package statsmon

import (
	"container/list"
	"fmt"
	"math"
	"time"

	"github.com/yangchenxing/foochow/stats"
)

type item struct {
	name     string
	item     stats.StatsItem
	upper    float64
	lower    float64
	callback MonitorCallback
}

type MonitorCallback func(string, float64, float64, float64, time.Duration)

var (
	tasks = make(map[time.Duration]*list.List)
)

func SetStatsItemMonitor(name string, upper, lower float64, interval time.Duration, callback MonitorCallback) error {
	statsItem := stats.GetStatsItem(name)
	if statsItem == nil {
		return fmt.Errorf("未知统计项: %s", name)
	}
	task := tasks[interval]
	if task == nil {
		task = list.New()
		tasks[interval] = task
		go monitor(interval)
	}
	task.PushBack(&item{
		name:     name,
		item:     statsItem,
		upper:    upper,
		lower:    lower,
		callback: callback,
	})
	return nil
}

func monitor(interval time.Duration) {
	for {
		time.Sleep(interval)
		for elem := tasks[interval].Front(); elem != nil; elem = elem.Next() {
			item := elem.Value.(*item)
			value := item.item.Get(interval)
			if !math.IsNaN(item.upper) && item.upper < value {
				item.callback(item.name, value, item.upper, item.lower, interval)
			} else if !math.IsNaN(item.lower) && item.lower > value {
				item.callback(item.name, value, item.upper, item.lower, interval)
			}
		}
	}
}
