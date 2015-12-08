package stats

import (
	"sync"
	"time"
)

type StatsItem interface {
	Get(time.Duration) float64
}

func NewCounter(name string) *Counter {
	rrdLock.Lock()
	defer rrdLock.Unlock()

	counter := &Counter{
		values: newData(),
	}
	registerItem(name, counter)
	return counter
}

type Counter struct {
	sync.Mutex
	values []float64
}

func (counter *Counter) Add(value float64) {
	counter.Lock()
	defer counter.Unlock()

	counter.values[current] += value
}

func (counter *Counter) Get(duration time.Duration) float64 {
	counter.Lock()
	defer counter.Unlock()

	count := int(duration.Seconds())
	if count > serialSize-1 {
		count = serialSize - 1
	}
	pos := (current - count + serialSize) % serialSize
	sum := float64(0)
	for i := 0; i < count; i++ {
		sum += counter.values[pos]
		pos = (pos + 1) % serialSize
	}
	return sum
}

func NewAverager(name, sumName, countName string) *Averager {
	averager := &Averager{
		sum:   NewCounter(sumName),
		count: NewCounter(countName),
	}
	registerItem(name, averager)
	return averager
}

type Averager struct {
	sync.Mutex
	sum   *Counter
	count *Counter
}

func (averager *Averager) Add(value float64) {
	averager.Lock()
	defer averager.Unlock()
	averager.sum.Add(value)
	averager.count.Add(1)
}

func (averager *Averager) Get(duration time.Duration) float64 {
	averager.Lock()
	defer averager.Unlock()
	sum := averager.sum.Get(duration)
	count := averager.count.Get(duration)
	if count == 0 {
		return 0
	}
	return sum / count
}

func (averager *Averager) GetAll(duration time.Duration) (value float64, sum float64, count uint64) {
	averager.Lock()
	defer averager.Unlock()
	sum = averager.sum.Get(duration)
	count = uint64(averager.count.Get(duration))
	if count > 0 {
		value = sum / float64(count)
	}
	return value, sum, count
}

func (averager *Averager) GetSum(duration time.Duration) float64 {
	averager.Lock()
	defer averager.Unlock()
	return averager.sum.Get(duration)
}

func (averager *Averager) GetCount(duration time.Duration) uint64 {
	averager.Lock()
	defer averager.Unlock()
	return uint64(averager.count.Get(duration))
}

func NewRatio(name, positiveName, totalName string) *Ratio {
	ratio := &Ratio{
		positive: NewCounter(positiveName),
		total:    NewCounter(totalName),
	}
	registerItem(name, ratio)
	return ratio
}

type Ratio struct {
	sync.Mutex
	positive *Counter
	total    *Counter
}

func (ratio *Ratio) Add(positive bool) {
	ratio.Lock()
	defer ratio.Unlock()
	if positive {
		ratio.positive.Add(1)
	}
	ratio.total.Add(1)
}

func (ratio *Ratio) AddPositive() {
	ratio.Lock()
	defer ratio.Unlock()
	ratio.positive.Add(1)
	ratio.total.Add(1)
}

func (ratio *Ratio) AddNegative() {
	ratio.Lock()
	defer ratio.Unlock()
	ratio.total.Add(1)
}

func (ratio *Ratio) Get(duration time.Duration) float64 {
	ratio.Lock()
	defer ratio.Unlock()
	positive := ratio.positive.Get(duration)
	total := ratio.total.Get(duration)
	if total == 0 {
		return 0
	}
	return positive / total
}

func (ratio *Ratio) GetAll(duration time.Duration) (value float64, positive uint64, total uint64) {
	ratio.Lock()
	defer ratio.Unlock()
	positive = uint64(ratio.positive.Get(duration))
	total = uint64(ratio.total.Get(duration))
	if total > 0 {
		value = float64(positive) / float64(total)
	}
	return
}

func (ratio *Ratio) GetPositive(duration time.Duration) uint64 {
	ratio.Lock()
	defer ratio.Unlock()
	return uint64(ratio.positive.Get(duration))
}

func (ratio *Ratio) GetTotal(duration time.Duration) uint64 {
	ratio.Lock()
	defer ratio.Unlock()
	return uint64(ratio.total.Get(duration))
}

func NewFrequency(name, counterName string) *Frequency {
	frequency := &Frequency{
		Counter: NewCounter(counterName),
	}
	registerItem(name, frequency)
	return frequency
}

type Frequency struct {
	*Counter
}

func (frequency *Frequency) Add() {
	frequency.Counter.Add(1)
}

func (frequency *Frequency) Get(duration time.Duration) float64 {
	return frequency.Counter.Get(duration) / duration.Seconds()
}

type Value float64

func NewValue(name string) *Value {
	value := new(Value)
	registerItem(name, value)
	return value
}

func (value *Value) Get(_ time.Duration) float64 {
	return float64(*value)
}

func (value *Value) Set(v float64) {
	*value = Value(v)
}
