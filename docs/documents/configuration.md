# ToughRADIUS Configuration

This configuration file is a YAML format configuration file of TOUGHRADIUS, which defines various settings of the TOUGHRADIUS server. Below is a detailed description of each section in the configuration file:

```yml
system:
  appid: ToughRADIUS # Application ID, used to identify the TOUGHRADIUS instance
  location: Asia/Shanghai # Time zone setting for the region where the server is located
  workdir: /var/toughradius # The working directory of TOUGHRADIUS, used to store logs, data files, etc.
  debug: true #Whether to enable debugging mode, more log information will be output when enabled.

web:
  host: 0.0.0.0 # The host address that the Web service monitors, 0.0.0.0 means monitoring all network interfaces
  port: 1816 # The port number that the Web service listens to
  tls_port: 1817 # The TLS encryption port number that the web service listens to
  secret: 9b6de5cc-0731-1203-xxtt-0f568ac9da37 # Web service key, used for encryption and other security-related operations

database:
  type: postgres #Database type, PostgreSQL is used here
  host: 127.0.0.1 # Host address of the database server
  port: 5432 #Port number of the database server
  name: toughradius_v8 # Database name
  user: postgres # Database username
  passwd: root # Database password
  max_conn: 100 # Maximum number of database connections
  idle_conn: 10 # Number of database idle connections
  debug: false # Whether to enable debugging mode for database operations

freeradius:
  enabled: true # Whether to enable FreeRADIUS integration
  host: 0.0.0.0 # The host address monitored by the FreeRADIUS service
  port: 1818 # The port number monitored by the FreeRADIUS service
  debug: true # Whether the FreeRADIUS service enables debugging mode

radiusd:
  enabled: true # Whether to enable the built-in RADIUS service
  host: 0.0.0.0 # The host address that the RADIUS service listens to
  auth_port: 1812 # The port number of the RADIUS authentication service
  acct_port: 1813 # Port number of RADIUS accounting service
  radsec_port: 2083 # Port number of RADIUS security service
  debug: true # Whether the RADIUS service enables debugging mode

tr069:
  host: 0.0.0.0 # TR069 service listening host address
  port: 1819 # TR069 service listening port number
  tls: false #Whether to enable TLS encryption, set to false here to disable it
  secret: 9b6de5cc-0731-1203-xxtt-0f568ac9da37 # TR069 service key
  debug: true # Whether the TR069 service enables debugging mode

mqtt:
  server: "" #MQTT server address, if you use MQTT you need to configure it
  client_id: "" # MQTT client ID
  username: "" # Username of MQTT service
  password: "" # Password for MQTT service
  debug: false # Whether the MQTT service enables debugging mode

logger:
  mode: development #Log mode, development means development mode
  console_enable: true # Whether to output logs on the console
  loki_enable: false # Whether to enable the Loki log aggregation system
  file_enable: true # Whether to enable file logs
  filename: /var/toughradius/toughradius.log # Path to the log file
  queue_size: 4096 # Log queue size
  loki_api: http://127.0.0.1:3100 # API address of Loki service
  loki_user: toughradius # Username for Loki service
  loki_pwd: toughradius # Password for Loki service
  loki_job: toughradius # Job name of Loki service
  metrics_storage: /var/toughradius/data/metrics #Metric data storage path
  metrics_history: 168 # Historical retention time of indicator data (hours)
```

Please note that some settings in this configuration file may need to be adjusted based on your actual environment. For example, the database username and password should be set to your actual credentials for the database,
TLS related settings should be adjusted depending on whether you use TLS, and MQTT and Loki configuration depends on whether you use these services. After modifying the configuration file,
It is usually necessary to restart the TOUGHRADIUS service for the changes to take effect.