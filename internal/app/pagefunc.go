package app

import (
	"time"
)

func (a *Application) GetTemplateFuncMap() map[string]interface{} {
	return map[string]interface{}{
		"pagever": func() int64 {
			if a.appConfig.System.Debug {
				return time.Now().Unix()
			} else {
				return int64(time.Now().Hour())
			}
		},
	}
}
