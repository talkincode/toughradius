## TR069 Certificate

```bash
# 1 Generate CA private key
test -f assets/ca.key || openssl genrsa -out assets/ca.key 4096
# 2 Generate CA certificate
test -f assets/ca.crt || openssl req -x509 -new -nodes -key assets/ca.key -days 3650 -out assets/ca.crt -subj \
"/C=CN/ST=Shanghai/O=toughradius/CN=ToughradiusCA/emailAddress=master@toughstruct.net"
# 3 Generate server private key
openssl genrsa -out assets/server.key 2048
# 4 Generate a certificate request file
openssl req -new -key assets/server.key -out assets/server.csr -subj \
"/C=CN/ST=Shanghai/O=toughradius/CN=*.toughstruct.net/emailAddress=master@toughstruct.net"
# 5 Generate a server certificate based on the CA's private key and the above certificate request file
openssl x509 -req -in assets/server.csr -CA assets/ca.crt -CAkey assets/ca.key -CAcreateserial -out assets/server.crt -days 7300
mv assets/server.key assets/cwmp.tls.key
mv assets/server.crt assets/cwmp.tls.crt
```

> It should be noted that the certificate prefix cwmp.tls is fixed, toughradius program will default to /var/toughradius/private/ directory, if there is no certificate file, it will create a default certificate file, default certificate file, CN=*.toughradius.net, May not work in your environment


## TR069 Preset template

The description format of presets is the standard YAML format, which can facilitate the use of various data structures. For example, online automatic initialization of new devices can be done using a set of presets

- oid  Can be set in scripts, factory settings, firmware configuration
- enabled Indicates whether to enable this task
- delay  For download tasks, the CPE can be delayed
- onfail  If it is defined as cancel, when the task fails, all unexecuted tasks defined by the description file will be canceled; when defined as ignore, unexecuted tasks will continue to be executed

The order of execution is  FactoryresetConfig -> FirmwareConfig -> Downloads ->Uploads -> SetParameterValues -> GetParameterNames,

In a set of preset tasks, only some operations can be set, if the preset is executed by the system scheduled task,factoryreset, firmwareconfig will be ignored

If the preset is performed by a scheduled system task (the time policy is set to `sys_scheduled`), then factoryreset, firmwareconfig are ignored in the set of preset tasks.


```yaml
# TR069 The default description format is the standard YAML format, which can facilitate the use of various data structures. For example, a set of presets can be used to complete the automatic initialization of new devices online

#	- oid  Can be set in scripts, factory settings, firmware configuration
#	- enabled Indicates whether to enable this task
#	- delay  For download tasks, the CPE can be delayed
#	- onfail  If it is defined as cancel, when the task fails, all unexecuted tasks defined by the description file will be canceled; when defined as ignore, unexecuted tasks will continue to be executed

# The order of execution is  FactoryresetConfig -> FirmwareConfig -> Downloads ->Uploads -> SetParameterValues -> GetParameterNames,

# If the preset is performed by a scheduled system task (the time policy is set to `sys_scheduled`), then factoryreset, firmwareconfig are ignored in the set of preset tasks.

# Factory settings script description, single definition
FactoryResetConfig:
  oid: "test_factory_reset_cfg"
  enabled: false
  delay: 0
  onfail: "ignore"

# Firmware configuration description
FirmwareConfig:
  oid: "test_firmware_cfg"
  enabled: false
  delay: 0
  onfail: "ignore"

# Regular download script, supports multiple sequential execution
Downloads:
  - oid: "test_download"
    enabled: true
    delay: 0
    onfail: "ignore"
  - oid: "test_download2"
    enabled: true
    delay: 0
    onfail: "ignore"

# Regular upload tasks that support multiple sequential executions
Uploads:
  - filetype: "2 VendorLog File"
    enabled: false
    onfail: "ignore"

# Set parameters, send multiple sets of parameters at one time
SetParameterValues:
  - name: "Device.DeviceInfo.X_MIKROTIK_SystemIdentity"
    type: "string"
    value: "TestRos"

# Get parameters, support multiple sequential execution
GetParameterValues:
  - "Device.DeviceInfo."

```