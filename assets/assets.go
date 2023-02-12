package assets

import (
	"embed"
)

//go:embed static
var StaticFs embed.FS

//go:embed templates
var TemplatesFs embed.FS

//go:embed build.txt
var BuildInfo string

//go:embed buildver.txt
var BuildVer string

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

//go:embed ca.crt
var CaCrt []byte
