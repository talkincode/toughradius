package installer

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/config"
	"gopkg.in/yaml.v3"
)

var installScript = `#!/bin/bash -x
mkdir -p /var/toughradius
chmod -R 755 /var/toughradius
install -m 755 {{binfile}} /usr/local/bin/toughradius
test -d /usr/lib/systemd/system || mkdir -p /usr/lib/systemd/system
cat>/usr/lib/systemd/system/toughradius.service<<EOF
[Unit]
Description=toughradius
After=network.target
StartLimitIntervalSec=0

[Service]
Restart=always
RestartSec=1
Environment=GODEBUG=x509ignoreCN=0
LimitNOFILE=65535
LimitNPROC=65535
User=root
ExecStart=/usr/local/bin/toughradius

[Install]
WantedBy=multi-user.target
EOF

chmod 600 /usr/lib/systemd/system/toughradius.service
systemctl enable toughradius && systemctl daemon-reload
`

func InitConfig(config *config.AppConfig) error {
	// config.NBI.JwtSecret = common.UUID()
	cfgstr, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile("/etc/toughradius.yml", cfgstr, 0644)
}

func Install() error {
	if !common.FileExists("/etc/toughradius.yml") {
		_ = InitConfig(config.DefaultAppConfig)
	}
	// Get the absolute path of the currently executing file
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	dir := filepath.Dir(path)
	binfile := filepath.Join(dir, "toughradius")
	installScript = strings.ReplaceAll(installScript, "{{binfile}}", binfile)
	_ = os.WriteFile("/tmp/toughradius_install.sh", []byte(installScript), 0755)

	// 创建用户&组
	if err := exec.Command("/bin/bash", "/tmp/toughradius_install.sh").Run(); err != nil {
		return err
	}

	return os.Remove("/tmp/toughradius_install.sh")
}

func Uninstall() {
	_ = os.Remove("/usr/lib/systemd/system/toughradius.service")
	_ = os.Remove("/usr/local/bin/toughradius")
}
