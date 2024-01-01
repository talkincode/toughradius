## ToughRADIUS Database Installation and Configuration Guide

This guide will provide a step-by-step process on how to install and configure the PostgreSQL database for the ToughRADIUS system on an Ubuntu 20.04 system.

## PostgreSQL Installation on Ubuntu 20.04

- Update System Packages

Before installing PostgreSQL, ensure that your system's package list and all installed packages are up-to-date.

```bash
sudo apt update
sudo apt upgrade
```

- Install PostgreSQL

To install PostgreSQL along with the pgAdmin tool, you can use the following command:

```bash
sudo apt install postgresql postgresql-contrib

```

- Confirm Installation

After the installation process is complete, you can check the status of the PostgreSQL service using the following command:

```bash
systemctl status postgresql

```
The service should be running and enabled to start on boot.


## Configuring the ToughRADIUS Database

Once you have PostgreSQL up and running, you can create a new user and database for ToughRADIUS.

- Switch to the PostgreSQL User

By default, a new system user named postgres is created during the installation of PostgreSQL. Switch to this user with the following command:

```bash
sudo su - postgres
```

- Enter the PostgreSQL Command Prompt

You can access the PostgreSQL command prompt by typing psql.

- Create a New User and Database

Run the following commands to create a new user toughradius with password toughradius, a database named toughradius, and grant all privileges on this database to the toughradius user:

```sql
CREATE USER toughradius WITH PASSWORD 'toughradius';
CREATE DATABASE toughradius WITH OWNER toughradius;
GRANT ALL PRIVILEGES ON DATABASE toughradius TO toughradius;

```

- Exit the PostgreSQL Prompt

To exit the PostgreSQL prompt, type \q.

- Exit the postgres User Shell

To switch back to your regular user, type exit.

## Docker Deployment of PostgreSQL

For deploying PostgreSQL using Docker, you can use Docker Compose with the following docker-compose.yml file:

```yaml
version: '3.1'

services:
  db:
    image: postgres
    restart: always
    environment:
      POSTGRES_USER: toughradius
      POSTGRES_PASSWORD: toughradius
      POSTGRES_DB: toughradius

```


To start the service, run docker-compose up -d.

> Note: Make sure you have Docker and Docker Compose installed on your system before running these commands.

## Important Information

ToughRADIUS adopts a mechanism where it will automatically create tables and initialize the database during the first startup. However, it's crucial to ensure that the database connection configuration is correct, otherwise, the automatic initialization will not be successful. For production environments, always double-check all configuration settings and test the database connection before launching the service.
