## toughradius  configuration


```yaml
system:
  appid: ToughRADIUS
  location: Asia/Shanghai
  workdir: /var/toughradius
  debug: true
web:
  host: 0.0.0.0
  port: 1816
  tls_port: 1817
  secret: 9b6de5cc-demo-1203-xxtt-0f568ac9da37
database:
  type: postgres
  host: 127.0.0.1
  port: 5432
  name: toughradius
  user: toughradius
  passwd: toughradius
  max_conn: 100
  idle_conn: 10
  debug: false
freeradius:
  enabled: true
  host: 0.0.0.0
  port: 1818
  debug: true
radiusd:
  enabled: true
  host: 0.0.0.0
  auth_port: 1812
  acct_port: 1813
  radsec_port: 2083
  debug: true
tr069:
  host: 0.0.0.0
  port: 1819
  tls: false
  secret: 9b6de5cc-0731-1203-xxtt-0f568ac9da37
  debug: true
mqtt:
  server: ""
  username: ""
  password: ""
  debug: false
logger:
  mode: development
  console_enable: true
  loki_enable: false
  file_enable: true
  filename: /var/toughradius/toughradius.log
  queue_size: 4096
  loki_api: http://127.0.0.1:3100
  loki_user: toughradius
  loki_pwd: toughradius
  loki_job: toughradius
  metrics_storage: /var/toughradius/data/metrics
  metrics_history: 168

```

## system

* appid: The identifier for TOUGHRADIUS application.
* location: The region and timezone where TOUGHRADIUS is running.
* workdir: The path to TOUGHRADIUS working directory.
* debug: Whether to enable debug mode for TOUGHRADIUS.

## web

* host: The binding host address for TOUGHRADIUS web service.
* port: The port number for TOUGHRADIUS web service.
* tls_port: The TLS port number for TOUGHRADIUS web service.
* secret: The secret key for TOUGHRADIUS web service.

## database

* type: The type of the database, which is PostgreSQL in this case.
* host: The host address of the database server.
* port: The port number of the database server.
* name: The name of the database.
* user: The username for the database.
* passwd: The password for the database.
* max_conn: The maximum number of database connections in the connection pool.
* idle_conn: The maximum number of idle connections in the connection pool.
* debug: Whether to enable debug mode for the database.

## freeradius

* enabled: Whether to enable the FreeRADIUS service.
* host: The binding host address for the FreeRADIUS service.
* port: The port number for the FreeRADIUS service.
* debug: Whether to enable debug mode for FreeRADIUS.

## radiusd: local radius service

* enabled: Whether to enable the RADIUS service.
* host: The binding host address for the RADIUS service.
* auth_port: The port number for RADIUS authentication.
* acct_port: The port number for RADIUS accounting.
* radsec_port: The port number for RADIUS with RadSec protocol.
* debug: Whether to enable debug mode for RADIUS.

## TR069

* host: The binding host address for the TR069 service.
* port: The port number for the TR069 service.
* tls: Whether to enable TLS encryption for TR069.
* secret: The secret key for TR069 communication.
* debug: Whether to enable debug mode for TR069.







