## Quick start

[中文](quickstart_cn)

### System environment dependencies

- Operating System：Support cross-platform deployment (Linux, Windows, MacOS, etc.)
- Database server: PostgreSQL 14+

### Database initialization

> Please do the installation and configuration yourself, first make sure your database server is running. [Database Setup](https://github.com/talkincode/toughradius/wiki/Database-Setup)

Running database creation scripts and creating dedicated users

```
CREATE USER toughradius WITH PASSWORD 'toughradius';
CREATE DATABASE toughradius WITH OWNER toughradius;
GRANT ALL PRIVILEGES ON DATABASE toughradius TO toughradius;
```

### Installation and Configuration

Let's take v8.0.4 as an example

Download the software distribution from [Releases Page](https://github.com/talkincode/toughradius/releases)

> If you have some development skills, you can compile your own version

```
curl https://github.com/talkincode/toughradius/releases/download/v8.0.6/toughradius_amd64 -O /tmp/toughradius

chmod +x /tmp/toughradius && /tmp/toughradius -install

```

Before proceeding, make sure that you have created the database and that the database server is running

Modifying configuration file [/etc/toughradius.yml](Configuration.md)


> The following installation method will download and build the latest toughradius version 

```bash
go install github.com/talkincode/toughradius/v8@latest

toughradius -install
```

Start the service with the following commands

    systemctl enable toughradius
    systemctl start toughradius

### Access the console

Open the browser and enter the URL: `http://your-ip:1816`

The default username and password are `admin/toughradius`