package radiusd

import (
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/app"
	"go.uber.org/zap"
	"layeh.com/radius"
)

func ListenRadiusAuthServer(appCtx app.AppContext, service *AuthService) error {
	cfg := appCtx.Config()
	if !cfg.Radiusd.Enabled {
		return nil
	}
	server := radius.PacketServer{
		Addr:               fmt.Sprintf("%s:%d", cfg.Radiusd.Host, cfg.Radiusd.AuthPort),
		Handler:            service,
		SecretSource:       service,
		InsecureSkipVerify: true,
	}

	zap.S().Infof("Starting Radius Auth server on %s", server.Addr)
	return server.ListenAndServe()
}

func ListenRadiusAcctServer(appCtx app.AppContext, service *AcctService) error {
	cfg := appCtx.Config()
	if !cfg.Radiusd.Enabled {
		return nil
	}
	server := radius.PacketServer{
		Addr:               fmt.Sprintf("%s:%d", cfg.Radiusd.Host, cfg.Radiusd.AcctPort),
		Handler:            service,
		SecretSource:       service,
		InsecureSkipVerify: true,
	}

	zap.S().Infof("Starting Radius Acct server on %s", server.Addr)
	return server.ListenAndServe()
}

func ListenRadsecServer(appCtx app.AppContext, service *RadsecService) error {
	cfg := appCtx.Config()
	if !cfg.Radiusd.Enabled {
		return nil
	}
	caCert := cfg.GetRadsecCaCertPath()
	serverCert := cfg.GetRadsecCertPath()
	serverKey := cfg.GetRadsecKeyPath()

	server := RadsecPacketServer{
		Addr:               fmt.Sprintf("%s:%d", cfg.Radiusd.Host, cfg.Radiusd.RadsecPort),
		Handler:            service,
		SecretSource:       service,
		InsecureSkipVerify: true,
		RadsecWorker:       cfg.Radiusd.RadsecWorker,
	}

	zap.S().Infof("Starting Radius Resec server on %s", server.Addr)
	err := server.ListenAndServe(caCert, serverCert, serverKey)
	if err != nil {
		zap.S().Errorf("Radius Resec server error: %s", err)
	}
	return err
}
