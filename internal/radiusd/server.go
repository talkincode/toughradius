package radiusd

import (
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/app"
	"go.uber.org/zap"
	"layeh.com/radius"
)

func ListenRadiusAuthServer(service *AuthService) error {
	if !app.GConfig().Radiusd.Enabled {
		return nil
	}
	server := radius.PacketServer{
		Addr:               fmt.Sprintf("%s:%d", app.GConfig().Radiusd.Host, app.GConfig().Radiusd.AuthPort),
		Handler:            service,
		SecretSource:       service,
		InsecureSkipVerify: true,
	}

	zap.S().Infof("Starting Radius Auth server on %s", server.Addr)
	return server.ListenAndServe()
}

func ListenRadiusAcctServer(service *AcctService) error {
	if !app.GConfig().Radiusd.Enabled {
		return nil
	}
	server := radius.PacketServer{
		Addr:               fmt.Sprintf("%s:%d", app.GConfig().Radiusd.Host, app.GConfig().Radiusd.AcctPort),
		Handler:            service,
		SecretSource:       service,
		InsecureSkipVerify: true,
	}

	zap.S().Infof("Starting Radius Acct server on %s", server.Addr)
	return server.ListenAndServe()
}

func ListenRadsecServer(service *RadsecService) error {
	if !app.GConfig().Radiusd.Enabled {
		return nil
	}
	caCert := app.GConfig().GetRadsecCaCertPath()
	serverCert := app.GConfig().GetRadsecCertPath()
	serverKey := app.GConfig().GetRadsecKeyPath()

	server := RadsecPacketServer{
		Addr:               fmt.Sprintf("%s:%d", app.GConfig().Radiusd.Host, app.GConfig().Radiusd.RadsecPort),
		Handler:            service,
		SecretSource:       service,
		InsecureSkipVerify: true,
		RadsecWorker:       app.GConfig().Radiusd.RadsecWorker,
	}

	zap.S().Infof("Starting Radius Resec server on %s", server.Addr)
	err := server.ListenAndServe(caCert, serverCert, serverKey)
	if err != nil {
		zap.S().Errorf("Radius Resec server error: %s", err)
	}
	return err
}
