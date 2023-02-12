package app

import (
	"strings"
	"time"

	"github.com/talkincode/toughradius/assets"
)

func (a *Application) GetTemplateFuncMap() map[string]interface{} {
	return map[string]interface{}{
		"pagever": func() int64 {
			return time.Now().Unix()
		},
		"buildver": func() string {
			bv := strings.TrimSpace(assets.BuildVer)
			if bv != "" {
				return bv
			}
			return "develop-" + time.Now().Format(time.RFC3339)
		},
		"zhlang": func() string {
			if a.GetTranslateLang() == ZhCN {
				return "1"
			}
			return "0"
		},
		"moontheme": func() string {
			theme := a.GetSystemTheme()
			if theme == "dark" {
				return "1"
			}
			return "0"
		},
		"theme": func() string {
			return a.GetSystemTheme()
		},
		"sys_config": func(name string) string {
			return a.GetSettingsStringValue("system", name)
		},
		"radius_config": func(name string) string {
			return a.GetSettingsStringValue("radius", name)
		},
		"tr069_config": func(name string) string {
			return a.GetSettingsStringValue("tr069", name)
		},
	}
}
