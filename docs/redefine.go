package docs

import (
	"os"
)

func Update() {
	newHost := os.Getenv("TOUGHRADIUS_SWAGGER_HOST")
	if newHost == "" {
		return
	}
	SwaggerInfo.Host = newHost
}
