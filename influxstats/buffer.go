package stats

import (
	"fmt"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

type Config struct {
	HTTPConfig client.HTTPConfig
	DB         string
	Interval   time.Duration
	Precision  string
}

type Transaction struct {
	items     ItemSet
	buffer    *Buffer
	submitted bool
}

func (transaction *Transaction) Submit() error {
	if transaction == nil {
		return nil
	}
	if transaction.submitted {
		return fmt.Errorf("重复提交统计事务")
	}
	transaction.buffer.transactions <- transaction
	transaction.submitted = true
	return nil
}

func (transaction *Transaction) AddFloat(measurement string, tags *Tags, name string, value float64) {
	if transaction != nil {
		transaction.items.AddFloat(measurement, tags, name, value)
	}
}

func (transaction *Transaction) AddInt(measurement string, tags *Tags, name string, value int64) {
	if transaction != nil {
		transaction.items.AddInt(measurement, tags, name, value)
	}
}

func (transaction *Transaction) SetFloat(measurement string, tags *Tags, name string, value float64) {
	if transaction != nil {
		transaction.items.SetFloat(measurement, tags, name, value)
	}
}

func (transaction *Transaction) SetInt(measurement string, tags *Tags, name string, value int64) {
	if transaction != nil {
		transaction.items.SetInt(measurement, tags, name, value)
	}
}

type MeasurementTags struct {
	measurement string
	tags        *Tags
}

func NewMeasurementTags(measurement string, tags ...string) *MeasurementTags {
	return &MeasurementTags{
		measurement: measurement,
		tags:        NewTags(tags...),
	}
}

func (mt *MeasurementTags) AddInt(tx *Transaction, name string, value int64) {
	tx.AddInt(mt.measurement, mt.tags, name, value)
}

func (mt *MeasurementTags) AddFloat(tx *Transaction, name string, value float64) {
	tx.AddFloat(mt.measurement, mt.tags, name, value)
}

func (mt *MeasurementTags) SetInt(tx *Transaction, name string, value int64) {
	tx.SetInt(mt.measurement, mt.tags, name, value)
}

func (mt *MeasurementTags) SetFloat(tx *Transaction, name string, value float64) {
	tx.SetFloat(mt.measurement, mt.tags, name, value)
}

type Buffer struct {
	client         client.Client
	db             string
	precision      string
	interval       time.Duration
	items          ItemSet
	errorCallback  func(error)
	submitCallback func(ItemSet)
	transactions   chan *Transaction
	submitTicker   chan time.Time
}

func NewBuffer(config Config, errorCallback func(error), submitCallback func(ItemSet)) (*Buffer, error) {
	client, err := client.NewHTTPClient(config.HTTPConfig)
	if err != nil {
		return nil, fmt.Errorf("创建InfluxDB客户端出错: %s", err.Error())
	}
	buffer := &Buffer{
		client:         client,
		db:             config.DB,
		interval:       config.Interval,
		items:          make(map[string]map[string]map[string]interface{}),
		transactions:   make(chan *Transaction, 64),
		submitTicker:   make(chan time.Time, 1),
		errorCallback:  errorCallback,
		submitCallback: submitCallback,
	}
	go func() {
		now := time.Now()
		next := now.Truncate(config.Interval).Add(config.Interval)
		time.Sleep(next.Sub(now))
		buffer.submitTicker <- next.Add(-1 * config.Interval)
		ticker := time.NewTicker(config.Interval)
		for t := range ticker.C {
			buffer.submitTicker <- t.Truncate(config.Interval).Add(-1 * config.Interval)
		}
	}()
	go buffer.readTransactions()
	return buffer, nil
}

func (buffer *Buffer) AddFloat(measurement string, tags *Tags, name string, value float64) {
	if buffer != nil {
		buffer.items.AddFloat(measurement, tags, name, value)
	}
}

func (buffer *Buffer) AddInt(measurement string, tags *Tags, name string, value int64) {
	if buffer != nil {
		buffer.items.AddInt(measurement, tags, name, value)
	}
}

func (buffer *Buffer) SetFloat(measurement string, tags *Tags, name string, value float64) {
	if buffer != nil {
		buffer.items.SetFloat(measurement, tags, name, value)
	}
}

func (buffer *Buffer) SetInt(measurement string, tags *Tags, name string, value int64) {
	if buffer != nil {
		buffer.items.SetInt(measurement, tags, name, value)
	}
}

func (buffer *Buffer) NewTransaction() *Transaction {
	if buffer == nil {
		return nil
	}
	return &Transaction{
		items:  make(map[string]map[string]map[string]interface{}),
		buffer: buffer,
	}
}

func (buffer *Buffer) onError(err error) {
	if buffer.errorCallback != nil {
		buffer.errorCallback(err)
	}
}

func (buffer *Buffer) submit(timestamp time.Time) error {
	defer buffer.reset()
	if buffer.submitCallback != nil {
		buffer.submitCallback(buffer.items)
	}
	points, err := client.NewBatchPoints(client.BatchPointsConfig{
		Precision: buffer.precision,
		Database:  buffer.db,
	})
	if err != nil {
		return fmt.Errorf("创建BatchPoints出错: %s", err.Error())
	}
	for measurement, tagItems := range buffer.items {
		for tag, fields := range tagItems {
			point, err := client.NewPoint(measurement, tagsCache[tag], fields, timestamp)
			if err != nil {
				return fmt.Errorf("创建Point出错: measurement=%q, tags=%v, fields=%v, timestamp=%s, error=%q",
					measurement, tagsCache[tag], fields, timestamp, err.Error())
			}
			points.AddPoint(point)
		}
	}
	go func() {
		if err := buffer.client.Write(points); err != nil {
			buffer.onError(fmt.Errorf("写入InfluxDB数据出错: %s", err.Error()))
		}
	}()
	return nil
}

func (buffer *Buffer) readTransactions() {
	for {
		select {
		case transaction := <-buffer.transactions:
			// fmt.Println("get transaction")
			buffer.add(transaction.items)
		case timestamp := <-buffer.submitTicker:
			// fmt.Println("get submit timestamp:", timestamp)
			buffer.submit(timestamp)
		}
	}
}

func (buffer *Buffer) add(items ItemSet) {
	for measurement, tagItems := range items {
		bufTagItems := buffer.items[measurement]
		if bufTagItems == nil {
			bufTagItems = make(map[string]map[string]interface{})
			buffer.items[measurement] = bufTagItems
		}
		for tag, fields := range tagItems {
			bufFields := bufTagItems[tag]
			if bufFields == nil {
				bufFields = make(map[string]interface{})
				bufTagItems[tag] = bufFields
			}
			for name, value := range fields {
				bufValue := bufFields[name]
				if bufValue == nil {
					bufFields[name] = value
				} else {
					switch v := value.(type) {
					case int64:
						bufFields[name] = bufValue.(int64) + v
					case float64:
						bufFields[name] = bufValue.(float64) + v
					}
				}
			}
		}
	}
}

func (buffer *Buffer) reset() {
	buffer.items = make(map[string]map[string]map[string]interface{})
}
