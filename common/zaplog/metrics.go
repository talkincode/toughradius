package zaplog

import (
	"encoding/json"
	slog "log"
	"time"

	"github.com/nakabonne/tstorage"
)

type metricsItem struct {
	Namespace string `json:"namespace"` // namespace
	Metrics   string `json:"metrics"`   // metrics
}

var emptyMetricsItem = metricsItem{}

type metricsWriter struct {
	tsdb        tstorage.Storage
	logChl      chan metricsItem
	stopProcess chan struct{}
}

func newMetricsWriter(queue int, metricpath string, retention time.Duration) *metricsWriter {
	mw := &metricsWriter{
		logChl:      make(chan metricsItem, 65535),
		stopProcess: make(chan struct{}),
	}
	var err error
	mw.tsdb, err = tstorage.NewStorage(
		tstorage.WithPartitionDuration(time.Hour),
		tstorage.WithTimestampPrecision(tstorage.Nanoseconds),
		tstorage.WithRetention(retention),
		tstorage.WithDataPath(metricpath),
		tstorage.WithWALBufferedSize(queue),
		tstorage.WithWriteTimeout(60*time.Second),
	)
	if err != nil {
		slog.Println(err.Error())
	}
	go mw.Start()
	return mw
}

func (c *metricsWriter) Write(p []byte) (int, error) {
	defer func() {
		if err := recover(); err != nil {
			err2, ok := err.(error)
			if ok {
				slog.Println(err2.Error())
			}
		}
	}()

	var item metricsItem
	err := json.Unmarshal(p, &item)
	if err != nil {
		return 0, err
	}

	_ = c.writeWithTimeout(c.logChl, item)

	return 0, nil
}

func (c *metricsWriter) readWithTimeout(ch <-chan metricsItem) (log metricsItem, err error) {
	select {
	case log = <-ch:
		return log, nil
	case <-time.After(time.Millisecond * 100):
		return emptyMetricsItem, readLogtimeout
	}
}

func (c *metricsWriter) writeWithTimeout(ch chan<- metricsItem, log metricsItem) (err error) {
	select {
	case ch <- log:
		return nil
	case <-time.After(time.Millisecond * 50):
		return writeLogtimeout
	}
}

func (c *metricsWriter) Start() {
	if c.tsdb == nil {
		return
	}
	timeC := time.NewTicker(time.Millisecond * 5000)

	processLog := func() {
		defer func() {
			if err := recover(); err != nil {
				slog.Println(err)
			}
		}()

		metrics := make(map[string]int64, 0)
		for {
			mitem, err := c.readWithTimeout(c.logChl)
			if err == readLogtimeout {
				break
			}
			if mitem == emptyMetricsItem {
				continue
			}
			if _, ok := metrics[mitem.Metrics]; ok {
				metrics[mitem.Metrics] += 1
			} else {
				metrics[mitem.Metrics] = 1
			}
		}

		var errcount int
		for k, v := range metrics {
			if err := c.tsdb.InsertRows([]tstorage.Row{
				{
					Metric: k,
					DataPoint: tstorage.DataPoint{
						Value:     float64(v),
						Timestamp: time.Now().Unix(),
					},
				},
			}); err != nil {
				errcount++
			}
		}

		if errcount > 0 {
			slog.Println("add timeseries data error total", errcount)
		}

	}

	for {
		select {
		case <-c.stopProcess:
			return
		case <-timeC.C:
			go processLog()
		}
	}
}

func (c *metricsWriter) Stop() {
	close(c.stopProcess)
	close(c.logChl)
}
