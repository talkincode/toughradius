#!/usr/bin/env python
#coding:utf-8
from toughradius.tools.secret import gen_secret

def echo_radiusd_cnf():
    return '''[DEFAULT]
debug = 0
tz = CST-8
secret = %s
ssl = 1
privatekey = /var/toughradius/privkey.pem
certificate = /var/toughradius/cacert.pem

[database]
dbtype = mysql
dburl = mysql://radiusd:root@127.0.0.1/toughradius?charset=utf8
echo = false
pool_size = 120
pool_recycle = 300

[radiusd]
acctport = 1813
adminport = 1815
authport = 1812
cache_timeout = 600
logfile = /var/toughradius/log/radiusd.log

[admin]
port = 1816
logfile = /var/toughradius/log/admin.log

[customer]
port = 1817
logfile = /var/toughradius/log/customer.log
'''%gen_secret(32)

def echo_privkey_pem():
    return '''-----BEGIN RSA PRIVATE KEY-----
MIIBPAIBAAJBAK+a5EAeEZFJdpwmMdgexCvE/x5HpsSvkyx+CFt9MDI8Gx9sXTsQ
hn+Satm4bNKq9+0yarGL1MoVoXCmzMkv++0CAwEAAQJBAJel139XeCxTmM54XYsZ
5qc11Gs9zVMFnL9Lh8QadEisGBoLNVGRKspVuR21pf9yWK1APJYtxeY+ElxTeN6v
frECIQDlXCN0ZLF2IBOUbOAEBnBEzYA19cnpktaD1EyeD1bpOwIhAMQAY3R+suNO
JE1MvE/g6ICAQVCDeiSW0JBUHbpXT5z3AiBakZqygHyPD7WLm76N+Fjm4lspc6hK
oqAwqGmk1JvWNwIhAJicyNPLV1S/4mpB5pq3v7FWrASZ6wAUYh8PL/qIw1evAiEA
sS5pdElUCN0d7/EdoOPBmEAJL7RHs6SjYEihK5ds4TQ=
-----END RSA PRIVATE KEY-----'''

def echo_cacert_pem():
    return '''-----BEGIN CERTIFICATE-----
MIIDTDCCAvagAwIBAgIJAMZsf8cd/CUeMA0GCSqGSIb3DQEBBQUAMIGiMQswCQYD
VQQGEwJDTjEOMAwGA1UECBMFSHVuYW4xETAPBgNVBAcTCENoYW5nc2hhMRgwFgYD
VQQKEw90b3VnaHJhZGl1cy5uZXQxFDASBgNVBAsTC3RvdWdocmFkaXVzMRgwFgYD
VQQDEw90b3VnaHJhZGl1cy5uZXQxJjAkBgkqhkiG9w0BCQEWF3N1cHBvcnRAdG91
Z2hyYWRpdXMubmV0MB4XDTE1MDMxODE2MTg1N1oXDTIwMTAyNTE2MTg1N1owgaIx
CzAJBgNVBAYTAkNOMQ4wDAYDVQQIEwVIdW5hbjERMA8GA1UEBxMIQ2hhbmdzaGEx
GDAWBgNVBAoTD3RvdWdocmFkaXVzLm5ldDEUMBIGA1UECxMLdG91Z2hyYWRpdXMx
GDAWBgNVBAMTD3RvdWdocmFkaXVzLm5ldDEmMCQGCSqGSIb3DQEJARYXc3VwcG9y
dEB0b3VnaHJhZGl1cy5uZXQwXDANBgkqhkiG9w0BAQEFAANLADBIAkEAr5rkQB4R
kUl2nCYx2B7EK8T/HkemxK+TLH4IW30wMjwbH2xdOxCGf5Jq2bhs0qr37TJqsYvU
yhWhcKbMyS/77QIDAQABo4IBCzCCAQcwHQYDVR0OBBYEFK9UjaxgsGyDZqfLEGUl
zYUhZqyzMIHXBgNVHSMEgc8wgcyAFK9UjaxgsGyDZqfLEGUlzYUhZqyzoYGopIGl
MIGiMQswCQYDVQQGEwJDTjEOMAwGA1UECBMFSHVuYW4xETAPBgNVBAcTCENoYW5n
c2hhMRgwFgYDVQQKEw90b3VnaHJhZGl1cy5uZXQxFDASBgNVBAsTC3RvdWdocmFk
aXVzMRgwFgYDVQQDEw90b3VnaHJhZGl1cy5uZXQxJjAkBgkqhkiG9w0BCQEWF3N1
cHBvcnRAdG91Z2hyYWRpdXMubmV0ggkAxmx/xx38JR4wDAYDVR0TBAUwAwEB/zAN
BgkqhkiG9w0BAQUFAANBAF2J27T8NnXptROTUx7IKU3MIBGvRqj6imtwjsus6fQU
GOLwDVfVEaqmv6YE6jg5ummEfeIcwUfkD5fLgrfRQ9s=
-----END CERTIFICATE-----'''

def echo_radiusd_script():
    return '''#!/bin/sh

### BEGIN INIT INFO
# Provides:               toughradius
# Required-Start:    $local_fs $remote_fs $network $syslog
# Required-Stop:         $local_fs $remote_fs $network $syslog
# Default-Start:         2 3 4 5
# Default-Stop:   0 1 6
# Short-Description: starts the toughradius daemon
# Description:     starts toughradius using start-stop-daemon
### END INIT INFO

usage () 
{
        cat <<EOF
Usage: $0 [OPTIONS]
  start              start toughradius 
  stop               stop toughradius 
  restart            restart toughradius, 
  upgrade            update toughradius version and restart
     
All other options are passed to the toughrad program.
EOF
        exit 1
}

start()
{
    toughctl --start all
}

stop()
{
    toughctl --stop all
}

restart()
{
    toughctl --restart all
}

upgrade()
{
    echo 'starting upgrade...' 
    pip install -U https://github.com/talkincode/ToughRADIUS/archive/stable.zip
    echo 'upgrade done'
}

case "$1" in

  help)
    usage
  ;;

  start)
    start
  ;;
  
  stop)
    stop
  ;;
    
  
  restart)
    restart
  ;;    
  
  upgrade)
    upgrade
  ;;  
  

  *)
   usage
  ;;

esac
exit 0
'''