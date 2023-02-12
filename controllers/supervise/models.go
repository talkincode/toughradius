package supervise

// SuperviseAction 远程管理Action
type SuperviseAction struct {
	Name    string `json:"name"` // 名称
	Type    string `json:"type"` // mkscript | extscript | snmp | tr069 | telnet | mikrotikapi
	Level string `json:"level"` // normal｜major
	Sid     string `json:"sid"`  // action ID
}
