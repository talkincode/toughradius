# ToughRADIUS Docker Quick Deployment

Here's the Docker deployment configuration for ToughRADIUS and PostgreSQL database, with detailed explanation as follows:

```yaml
version: "3"
services:
  pgdb:
    image: timescale/timescaledb:latest-pg14
    container_name: "pgdb"
    ports:
      - "127.0.0.1:5432:5432"
    environment:
      POSTGRES_DB: toughradius
      POSTGRES_USER: toughradius
      POSTGRES_PASSWORD: toughradius
    volumes:
      - pgdb-volume:/var/lib/postgresql/data
    networks:
      toughradius_network:

  toughradius:
    depends_on:
      - 'pgdb'
    image: talkincode/toughradius:latest
    container_name: "toughradius"
    restart: always
    ports:
      - "1816:1816"
      - "1818:1818"
      - "1819:1819"
      - "2083:2083"
      - "1812:1812/udp"
      - "1813:1813/udp"
      - "1914:1914/udp"
    volumes:
      - toughradius-volume:/var/toughradius
    environment:
      - GODEBUG=x509ignoreCN=0
      - TOUGHRADIUS_SYSTEM_DEBUG=off
      - TOUGHRADIUS_DB_HOST=pgdb
      - TOUGHRADIUS_DB_NAME=toughradius
      - TOUGHRADIUS_DB_USER=toughradius
      - TOUGHRADIUS_DB_PWD=toughradius
      - TOUGHRADIUS_RADIUS_DEBUG=off
      - TOUGHRADIUS_RADIUS_ENABLED=on
      - TOUGHRADIUS_TR069_WEB_TLS=on
      - TOUGHRADIUS_LOKI_ENABLE=false
      - TOUGHRADIUS_LOGGER_MODE=production
      - TOUGHRADIUS_LOGGER_FILE_ENABLE=true
    networks:
      toughradius_network:

networks:
  toughradius_network:

volumes:
  pgdb-volume:
  toughradius-volume:
```

## Services:

* pgdb: This is the service definition for the PostgresSQL database. We're using the TimescaleDB image, which is a PostgreSQL database optimized for time-series data. This service maps the container's port 5432 to the host's port 5432 for external access.

* toughradius: This is the service definition for ToughRADIUS. This service depends on 'pgdb', meaning it will only start once 'pgdb' service is fully up and running. It uses the image 'ca17/toughradius:latest'.

## Port Mapping:

* "1816:1816": Web management port
* "1818:1818": FreeRADIUS API Service Port
* "1819:1819": TR069 Service port
* "2083:2083": Radsec service port
* "1812:1812/udp": RADIUS Authentication port, using UDP protocol
* "1813:1813/udp": RADIUS Accounting port, using UDP protocol

## Volumes:

* pgdb-volume: This volume is mounted to /var/lib/postgresql/data in the PostgresSQL database container, and it's used to store database data.
* toughradius-volume: This volume is mounted to /var/toughradius in the ToughRADIUS container, and it's used to store ToughRADIUS data.

##  Networks:

* toughradius_network: All the services are linked to this network, ensuring communication between containers.
* Environment variables are used to configure various parameters of ToughRADIUS, such as database connection information, log settings, etc.

Once you've created this Docker Compose file, you can use the docker-compose up command to start all services.