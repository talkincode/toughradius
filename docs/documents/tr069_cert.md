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

