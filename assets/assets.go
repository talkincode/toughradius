package assets

import (
	"embed"
	"regexp"
)

//go:embed static
var StaticFs embed.FS

//go:embed templates
var TemplatesFs embed.FS

//go:embed buildinfo.txt
var BuildInfo string

//go:embed menu-admin.json
var AdminMenudata []byte

//go:embed menu-opr.json
var OprMenudata []byte

//go:embed pgdump_script.sh
var PgdumpShell string

//go:embed cwmp.tls.crt
var CwmpCert []byte

//go:embed cwmp.tls.key
var CwmpKey []byte

//go:embed ca.key
var CaKey []byte

//go:embed ca.crt
var CaCrt []byte

//go:embed tr069_mikrotik.rsc
var Tr069Mikrotik string

//go:embed tr069_preset.yml
var Tr069PresetTemplate string

var defaultBuildVer = "Latest Build 2023"

func BuildVersion() string {
	re, err := regexp.Compile(`BuildVersion=(.+?)\n`)
	if err != nil {
		return defaultBuildVer
	}
	match := re.FindStringSubmatch(BuildInfo)

	if len(match) > 0 {
		return match[1]
	}
	return defaultBuildVer
}
