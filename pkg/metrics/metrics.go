package metrics

import (
	"path"
	"time"

	"github.com/nakabonne/tstorage"
)

var globalTSDB tstorage.Storage

// InitMetrics initializes the time series database for metrics
// Uses convention: dataPath = workdir/data/metrics, retention = 7 days (168 hours)
func InitMetrics(workdir string) error {
	dataPath := path.Join(workdir, "data", "metrics")
	retention := time.Hour * 24 * 7 // 7 days

	var err error
	globalTSDB, err = tstorage.NewStorage(
		tstorage.WithPartitionDuration(time.Hour),
		tstorage.WithTimestampPrecision(tstorage.Nanoseconds),
		tstorage.WithRetention(retention),
		tstorage.WithDataPath(dataPath),
		tstorage.WithWALBufferedSize(4096),
		tstorage.WithWriteTimeout(60*time.Second),
	)
	return err
}

// GetTSDB returns the global time series database instance
func GetTSDB() tstorage.Storage {
	return globalTSDB
}

// Close closes the time series database
func Close() error {
	if globalTSDB != nil {
		return globalTSDB.Close()
	}
	return nil
}
