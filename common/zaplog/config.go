package zaplog

type LogConfig struct {
	Mode           string
	ConsoleEnable  bool
	LokiEnable     bool
	FileEnable     bool
	Filename       string
	LokiApi        string
	LokiUser       string
	LokiPwd        string
	LokiJob        string
	QueueSize      int
	MetricsStorage string
	MetricsHistory int
}
