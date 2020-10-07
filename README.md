# About TOUGHRADIUS 


                 )                        )   (                (       (              (
      *   )   ( /(            (        ( /(   )\ )     (       )\ )    )\ )           )\ )              (
    ` )  /(   )\())      (    )\ )     )\()) (()/(     )\     (()/(   (()/(      (   (()/(     (   (    )\ )
     ( )(_)) ((_)\       )\  (()/(    ((_)\   /(_)) ((((_)(    /(_))   /(_))     )\   /(_))    )\  )\  (()/(
    (_(_())    ((_)   _ ((_)  /(_))_   _((_) (_))    )\ _ )\  (_))_   (_))    _ ((_) (_))     ((_)((_)  /(_))
    |_   _|   / _ \  | | | | (_)) __| | || | | _ \   (_)_\(_)  |   \  |_ _|  | | | | / __|    \ \ / /  (_) /
      | |    | (_) | | |_| |   | (_ | | __ | |   /    / _ \    | |) |  | |   | |_| | \__ \     \ V /    / _ \
      |_|     \___/   \___/     \___| |_||_| |_|_\   /_/ \_\   |___/  |___|   \___/  |___/      \_/     \___/

                                              /)
        _   __   __   _    _/_  ___       _  (/   _  _/_  __      _ _/_      _  ______
        (_(/ (_(/ (_(/  .  (__ (_) (_(_  (_/_/ )_/_)_(__ / (_(_(_(__(__  .  (__(_) // (_
                                        .-/
                                       (_/

[中文](README_CN.md)

TOUGHRADIUS is an open source Radius service software that supports standard RADIUS protocol (RFC 2865, RFC 2866) and provides a complete AAA implementation. It supports flexible policy management, supports all major access devices and easily extends with rich billing policy support.

Redeveloped from version 6.x onwards, based on the Java language. A high-performance RADIUS processing engine is provided, along with a simple and easy-to-use web management interface that is easy to use.

TOUGHRADIUS is similar in functionality to freeRADIUS, but it is simpler to use and easier to develop by extension.

## Links

- [Home](https://www.toughradius.net/)
- [TOUGHRADIUS WIKI Documentation](https://github.com/talkincode/ToughRADIUS/wiki)
- [GUI Test Tool](https://github.com/jamiesun/RadiusTester)
- [Command line testing tool](https://github.com/talkincode/JRadiusTester)
- [Commercial support](https://www.toughstruct.net)

## Quick start

### System environment dependency

- Operating System：Support cross-platform deployment (Linux, Windows, MacOS, etc.)
- java version: 1.8 or higher
- Database server: MySQL/MariaDB

### Database initialization

> Please do the installation and configuration yourself, first make sure your database server is running.

Running database creation scripts and creating dedicated users

    create database toughradius DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
    GRANT ALL ON toughradius.* TO raduser@'127.0.0.1' IDENTIFIED BY 'radpwd' WITH GRANT OPTION;FLUSH PRIVILEGES;

Creating Database Tables

    create table if not exists tr_bras
    (
        id bigint auto_increment primary key,
        identifier varchar(128) null,
        name varchar(64) not null,
        ipaddr varchar(32) null,
        vendor_id varchar(32) not null,
        portal_vendor varchar(32) not null,
        secret varchar(64) not null,
        coa_port int not null,
        ac_port int not null,
        auth_limit int null,
        acct_limit int null,
        status enum('enabled', 'disabled') null,
        remark varchar(512) null,
        create_time datetime not null
    );

    create index ix_tr_bras_identifier on tr_bras (identifier);
    
    create index ix_tr_bras_ipaddr on tr_bras (ipaddr);
    
    create table if not exists tr_config
    (
        id bigint auto_increment primary key,
        type varchar(32) not null,
        name varchar(128) not null,
        value varchar(255) null,
        remark varchar(255) null
    );
    
    create table if not exists tr_subscribe
    (
        id bigint auto_increment primary key,
        node_id bigint default 0 not null,
        subscriber varchar(32) null,
        realname varchar(32) null,
        password varchar(128) not null,
        domain varchar(128) null,
        addr_pool varchar(128) null,
        policy varchar(512) null,
        is_online int null,
        active_num int null,
        bind_mac tinyint(1) null,
        bind_vlan tinyint(1) null,
        ip_addr varchar(32) null,
        mac_addr varchar(32) null,
        in_vlan int null,
        out_vlan int null,
        up_rate bigint null,
        down_rate bigint null,
        up_peak_rate bigint null,
        down_peak_rate bigint null,
        up_rate_code varchar(32) null,
        down_rate_code varchar(32) null,
        status enum('enabled', 'disabled') null,
        remark varchar(512) null,
        begin_time datetime not null,
        expire_time datetime not null,
        create_time datetime not null,
        update_time datetime null
    );
    
    create index ix_tr_subscribe_create_time
        on tr_subscribe (create_time);
    
    create index ix_tr_subscribe_expire_time
        on tr_subscribe (expire_time);
    
    create index ix_tr_subscribe_status
        on tr_subscribe (status);
    
    create index ix_tr_subscribe_subscriber
        on tr_subscribe (subscriber);
    
    create index ix_tr_subscribe_update_time
        on tr_subscribe (update_time);
    

Inserting test data

    INSERT INTO toughradius.tr_bras
    (identifier, name, ipaddr, vendor_id, portal_vendor,secret, coa_port,ac_port, auth_limit, acct_limit, STATUS, remark, create_time)
    VALUES ('radius-tester', 'radius-tester', '127.0.0.1', '14988',"cmccv1", 'secret', 3799,2000, 1000, 1000, NULL, '0', '2019-03-01 14:07:46');

    INSERT INTO toughradius.tr_subscribe
    (node_id,  subscriber, realname, password, domain, addr_pool, policy, is_online, active_num,
     bind_mac, bind_vlan, ip_addr, mac_addr, in_vlan, out_vlan, up_rate, down_rate, up_peak_rate, 
     down_peak_rate, up_rate_code,down_rate_code, status, remark, begin_time, expire_time, create_time, update_time)
    VALUES (0, 'test01', '', '888888',  null, null, null, null, 10, 0, 0, '', '', 0, 0, 10.000, 10.000, 100.000, 100.000,
            '10', '10', 'enabled', '', '2019-03-01 14:13:02', '2019-03-01 14:13:00', '2019-03-01 14:12:59', '2019-03-01 14:12:56');

### Run docker container
    
    export RADIUS_DBURL="jdbc:mysql://172.17.0.1:3306/toughradius?serverTimezone=Asia/Shanghai&useUnicode=true&characterEncoding=utf-8&allowMultiQueries=true"
    export RADIUS_DBUSER=raduser
    export RADIUS_DBPWD=radpwd
    
    docker run --name toughradius -d \
    -v /tradiusdata/vardata:/var/toughradius \
    --env RADIUS_DBURL \
    --env RADIUS_DBUSER \
    --env RADIUS_DBPWD \
    -p 1816:1816/tcp \
    -p 1812:1812/udp \
    -p 1813:1813/udp \
    talkincode/toughradius:latest
    
> [More references to environmental variables](https://github.com/talkincode/ToughRADIUS/wiki/docker_related)     
>
> [Run with docker-compose](https://github.com/talkincode/ToughRADIUS/wiki/docker_related)
            
### Run the main program

    java -jar -Xms256M -Xmx1024M /opt/toughradius-latest.jar --spring.profiles.active=prod
    
> Note the path to the jar file (toughradius-latest.jar).

### Linux systemd service configuration

/opt/application-prod.properties

    # web access port
    server.port = 1816
    
    # If https is enabled, just cancel the following comment
    #server.security.require-ssl=true
    #server.ssl.key-store-type=PKCS12
    #server.ssl.key-store=classpath:toughradius.p12
    #server.ssl.key-store-password=toughstruct
    #server.ssl.key-alias=toughradius
    
    # Log configuration, either logback-prod.xml or logback-dev.xml, logging directory /var/toughradius/logs
    logging.config=classpath:logback-prod.xml
    
    # Database configuration
    spring.datasource.url=${RADIUS_DBURL:jdbc:mysql://127.0.0.1:3306/toughradius?serverTimezone=Asia/Shanghai&useUnicode=true&characterEncoding=utf-8&allowMultiQueries=true}
    spring.datasource.username=${RADIUS_DBUSER:raduser}
    spring.datasource.password=${RADIUS_DBPWD:radpwd}
    spring.datasource.max-active=${RADIUS_DBPOOL:120}
    spring.datasource.driver-class-name=com.mysql.cj.jdbc.Driver

/usr/lib/systemd/system/toughradius.service

    [Unit]
    Description=toughradius
    After=syslog.target
    
    [Service]
    WorkingDirectory=/opt
    User=root
    LimitNOFILE=65535
    LimitNPROC=65535
    Type=simple
    ExecStart=/usr/bin/java -server -jar -Xms256M -Xmx1024M /opt/toughradius-latest.jar  --spring.profiles.active=prod
    SuccessExitStatus=143
    
    [Install]
    WantedBy=multi-user.target

> If you understand spring systemd and configuration principles, you can modify it according to your actual needs.

Start the service with the following commands

    systemctl enable toughradius
    systemctl start toughradius
    
