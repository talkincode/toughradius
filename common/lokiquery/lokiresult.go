package lokiquery

import (
	"time"
)

type LokiQueryResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream struct {
				LevelExtracted string    `json:"level_extracted"`
				Msg            string    `json:"msg"`
				Ts             time.Time `json:"ts"`
				Caller         string    `json:"caller"`
				Job            string    `json:"job"`
				Level          string    `json:"level"`
				Namespace      string    `json:"namespace,omitempty"`
				Metrics        string    `json:"metrics,omitempty"`
				Username       string    `json:"username,omitempty"`
				Result         string    `json:"result,omitempty"`
				Nasip          string    `json:"nasip,omitempty"`
			} `json:"stream"`
			Values [][]string `json:"values"`
		} `json:"result"`
		Stats struct {
			Summary struct {
				BytesProcessedPerSecond int     `json:"bytesProcessedPerSecond"`
				LinesProcessedPerSecond int     `json:"linesProcessedPerSecond"`
				TotalBytesProcessed     int     `json:"totalBytesProcessed"`
				TotalLinesProcessed     int     `json:"totalLinesProcessed"`
				ExecTime                float64 `json:"execTime"`
				QueueTime               float64 `json:"queueTime"`
				Subqueries              int     `json:"subqueries"`
				TotalEntriesReturned    int     `json:"totalEntriesReturned"`
			} `json:"summary"`
			Querier struct {
				Store struct {
					TotalChunksRef        int `json:"totalChunksRef"`
					TotalChunksDownloaded int `json:"totalChunksDownloaded"`
					ChunksDownloadTime    int `json:"chunksDownloadTime"`
					Chunk                 struct {
						HeadChunkBytes    int `json:"headChunkBytes"`
						HeadChunkLines    int `json:"headChunkLines"`
						DecompressedBytes int `json:"decompressedBytes"`
						DecompressedLines int `json:"decompressedLines"`
						CompressedBytes   int `json:"compressedBytes"`
						TotalDuplicates   int `json:"totalDuplicates"`
					} `json:"chunk"`
				} `json:"store"`
			} `json:"querier"`
			Ingester struct {
				TotalReached       int `json:"totalReached"`
				TotalChunksMatched int `json:"totalChunksMatched"`
				TotalBatches       int `json:"totalBatches"`
				TotalLinesSent     int `json:"totalLinesSent"`
				Store              struct {
					TotalChunksRef        int `json:"totalChunksRef"`
					TotalChunksDownloaded int `json:"totalChunksDownloaded"`
					ChunksDownloadTime    int `json:"chunksDownloadTime"`
					Chunk                 struct {
						HeadChunkBytes    int `json:"headChunkBytes"`
						HeadChunkLines    int `json:"headChunkLines"`
						DecompressedBytes int `json:"decompressedBytes"`
						DecompressedLines int `json:"decompressedLines"`
						CompressedBytes   int `json:"compressedBytes"`
						TotalDuplicates   int `json:"totalDuplicates"`
					} `json:"chunk"`
				} `json:"store"`
			} `json:"ingester"`
			Cache struct {
				Chunk struct {
					EntriesFound     int `json:"entriesFound"`
					EntriesRequested int `json:"entriesRequested"`
					EntriesStored    int `json:"entriesStored"`
					BytesReceived    int `json:"bytesReceived"`
					BytesSent        int `json:"bytesSent"`
					Requests         int `json:"requests"`
				} `json:"chunk"`
				Index struct {
					EntriesFound     int `json:"entriesFound"`
					EntriesRequested int `json:"entriesRequested"`
					EntriesStored    int `json:"entriesStored"`
					BytesReceived    int `json:"bytesReceived"`
					BytesSent        int `json:"bytesSent"`
					Requests         int `json:"requests"`
				} `json:"index"`
				Result struct {
					EntriesFound     int `json:"entriesFound"`
					EntriesRequested int `json:"entriesRequested"`
					EntriesStored    int `json:"entriesStored"`
					BytesReceived    int `json:"bytesReceived"`
					BytesSent        int `json:"bytesSent"`
					Requests         int `json:"requests"`
				} `json:"result"`
			} `json:"cache"`
		} `json:"stats"`
	} `json:"data"`
}

type LokiMetricResult struct {
	Data struct {
		Result []struct {
			Metric struct {
				Job   string `json:"job"`
				Level string `json:"level,omitempty"`
			} `json:"metric"`
			Values [][]interface{} `json:"values"`
		} `json:"result"`
		ResultType string `json:"resultType"`
		Stats      struct {
			Cache struct {
				Chunk struct {
					BytesReceived    int `json:"bytesReceived"`
					BytesSent        int `json:"bytesSent"`
					EntriesFound     int `json:"entriesFound"`
					EntriesRequested int `json:"entriesRequested"`
					EntriesStored    int `json:"entriesStored"`
					Requests         int `json:"requests"`
				} `json:"chunk"`
				Index struct {
					BytesReceived    int `json:"bytesReceived"`
					BytesSent        int `json:"bytesSent"`
					EntriesFound     int `json:"entriesFound"`
					EntriesRequested int `json:"entriesRequested"`
					EntriesStored    int `json:"entriesStored"`
					Requests         int `json:"requests"`
				} `json:"index"`
				Result struct {
					BytesReceived    int `json:"bytesReceived"`
					BytesSent        int `json:"bytesSent"`
					EntriesFound     int `json:"entriesFound"`
					EntriesRequested int `json:"entriesRequested"`
					EntriesStored    int `json:"entriesStored"`
					Requests         int `json:"requests"`
				} `json:"result"`
			} `json:"cache"`
			Ingester struct {
				Store struct {
					Chunk struct {
						CompressedBytes   int `json:"compressedBytes"`
						DecompressedBytes int `json:"decompressedBytes"`
						DecompressedLines int `json:"decompressedLines"`
						HeadChunkBytes    int `json:"headChunkBytes"`
						HeadChunkLines    int `json:"headChunkLines"`
						TotalDuplicates   int `json:"totalDuplicates"`
					} `json:"chunk"`
					ChunksDownloadTime    int `json:"chunksDownloadTime"`
					TotalChunksDownloaded int `json:"totalChunksDownloaded"`
					TotalChunksRef        int `json:"totalChunksRef"`
				} `json:"store"`
				TotalBatches       int `json:"totalBatches"`
				TotalChunksMatched int `json:"totalChunksMatched"`
				TotalLinesSent     int `json:"totalLinesSent"`
				TotalReached       int `json:"totalReached"`
			} `json:"ingester"`
			Querier struct {
				Store struct {
					Chunk struct {
						CompressedBytes   int `json:"compressedBytes"`
						DecompressedBytes int `json:"decompressedBytes"`
						DecompressedLines int `json:"decompressedLines"`
						HeadChunkBytes    int `json:"headChunkBytes"`
						HeadChunkLines    int `json:"headChunkLines"`
						TotalDuplicates   int `json:"totalDuplicates"`
					} `json:"chunk"`
					ChunksDownloadTime    int `json:"chunksDownloadTime"`
					TotalChunksDownloaded int `json:"totalChunksDownloaded"`
					TotalChunksRef        int `json:"totalChunksRef"`
				} `json:"store"`
			} `json:"querier"`
			Summary struct {
				BytesProcessedPerSecond int     `json:"bytesProcessedPerSecond"`
				ExecTime                float64 `json:"execTime"`
				LinesProcessedPerSecond int     `json:"linesProcessedPerSecond"`
				QueueTime               float64 `json:"queueTime"`
				Subqueries              int     `json:"subqueries"`
				TotalBytesProcessed     int     `json:"totalBytesProcessed"`
				TotalEntriesReturned    int     `json:"totalEntriesReturned"`
				TotalLinesProcessed     int     `json:"totalLinesProcessed"`
			} `json:"summary"`
		} `json:"stats"`
	} `json:"data"`
	Status string `json:"status"`
}

type LokiSumRateResult struct {
	Data struct {
		Result []struct {
			Metric struct {
			} `json:"metric"`
			Values [][]interface{} `json:"values"`
		} `json:"result"`
		ResultType string `json:"resultType"`
		Stats      struct {
			Cache struct {
				Chunk struct {
					BytesReceived    int `json:"bytesReceived"`
					BytesSent        int `json:"bytesSent"`
					EntriesFound     int `json:"entriesFound"`
					EntriesRequested int `json:"entriesRequested"`
					EntriesStored    int `json:"entriesStored"`
					Requests         int `json:"requests"`
				} `json:"chunk"`
				Index struct {
					BytesReceived    int `json:"bytesReceived"`
					BytesSent        int `json:"bytesSent"`
					EntriesFound     int `json:"entriesFound"`
					EntriesRequested int `json:"entriesRequested"`
					EntriesStored    int `json:"entriesStored"`
					Requests         int `json:"requests"`
				} `json:"index"`
				Result struct {
					BytesReceived    int `json:"bytesReceived"`
					BytesSent        int `json:"bytesSent"`
					EntriesFound     int `json:"entriesFound"`
					EntriesRequested int `json:"entriesRequested"`
					EntriesStored    int `json:"entriesStored"`
					Requests         int `json:"requests"`
				} `json:"result"`
			} `json:"cache"`
			Ingester struct {
				Store struct {
					Chunk struct {
						CompressedBytes   int `json:"compressedBytes"`
						DecompressedBytes int `json:"decompressedBytes"`
						DecompressedLines int `json:"decompressedLines"`
						HeadChunkBytes    int `json:"headChunkBytes"`
						HeadChunkLines    int `json:"headChunkLines"`
						TotalDuplicates   int `json:"totalDuplicates"`
					} `json:"chunk"`
					ChunksDownloadTime    int `json:"chunksDownloadTime"`
					TotalChunksDownloaded int `json:"totalChunksDownloaded"`
					TotalChunksRef        int `json:"totalChunksRef"`
				} `json:"store"`
				TotalBatches       int `json:"totalBatches"`
				TotalChunksMatched int `json:"totalChunksMatched"`
				TotalLinesSent     int `json:"totalLinesSent"`
				TotalReached       int `json:"totalReached"`
			} `json:"ingester"`
			Querier struct {
				Store struct {
					Chunk struct {
						CompressedBytes   int `json:"compressedBytes"`
						DecompressedBytes int `json:"decompressedBytes"`
						DecompressedLines int `json:"decompressedLines"`
						HeadChunkBytes    int `json:"headChunkBytes"`
						HeadChunkLines    int `json:"headChunkLines"`
						TotalDuplicates   int `json:"totalDuplicates"`
					} `json:"chunk"`
					ChunksDownloadTime    int `json:"chunksDownloadTime"`
					TotalChunksDownloaded int `json:"totalChunksDownloaded"`
					TotalChunksRef        int `json:"totalChunksRef"`
				} `json:"store"`
			} `json:"querier"`
			Summary struct {
				BytesProcessedPerSecond int     `json:"bytesProcessedPerSecond"`
				ExecTime                float64 `json:"execTime"`
				LinesProcessedPerSecond int     `json:"linesProcessedPerSecond"`
				QueueTime               float64 `json:"queueTime"`
				Subqueries              int     `json:"subqueries"`
				TotalBytesProcessed     int     `json:"totalBytesProcessed"`
				TotalEntriesReturned    int     `json:"totalEntriesReturned"`
				TotalLinesProcessed     int     `json:"totalLinesProcessed"`
			} `json:"summary"`
		} `json:"stats"`
	} `json:"data"`
	Status string `json:"status"`
}
