package toughradius

import (
	"fmt"
	"os"
	"path"

	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/assets"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/zaplog/log"
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

	log.Infof("Starting Radius Auth server on %s", server.Addr)
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

	log.Infof("Starting Radius Acct server on %s", server.Addr)
	return server.ListenAndServe()
}

func ListenRadsecServer(service *RadsecService) error {
	if !app.GConfig().Radiusd.Enabled {
		return nil
	}
	caCert := path.Join(app.GConfig().System.Workdir, "private/ca.crt")
	serverCert := path.Join(app.GConfig().System.Workdir, "private/radsec.tls.crt")
	serverKey := path.Join(app.GConfig().System.Workdir, "private/radsec.tls.key")
	if !common.FileExists(caCert) {
		os.WriteFile(caCert, assets.CaCrt, 0644)
	}
	if !common.FileExists(serverCert) {
		os.WriteFile(serverCert, assets.CwmpCert, 0644)
	}
	if !common.FileExists(serverKey) {
		os.WriteFile(serverKey, assets.CwmpKey, 0644)
	}

	server := RadsecPacketServer{
		Addr:               fmt.Sprintf("%s:%d", app.GConfig().Radiusd.Host, app.GConfig().Radiusd.RadsecPort),
		Handler:            service,
		SecretSource:       service,
		InsecureSkipVerify: true,
		RadsecWorker:       app.GConfig().Radiusd.RadsecWorker,
	}

	log.Infof("Starting Radius Resec server on %s", server.Addr)
	err := server.ListenAndServe(caCert, serverCert, serverKey)
	if err != nil {
		log.Errorf("Radius Resec server error: %s", err)
	}
	return err
}
