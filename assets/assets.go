package assets

import (
	_ "embed"
)

//go:embed ca.key
var CaKey []byte

//go:embed ca.crt
var CaCrt []byte

//go:embed radsec.tls.crt
var RadsecCert []byte

//go:embed radsec.tls.key
var RadsecKey []byte
