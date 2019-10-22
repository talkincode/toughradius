# 关于 TOUGHRADIUS 


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


TOUGHRADIUS 是一个开源的Radius服务软件，支持标准RADIUS协议（RFC 2865, RFC 2866），提供完整的AAA实现。支持灵活的策略管理，支持各种主流接入设备并轻松扩展，具备丰富的计费策略支持。

至 6.x 版本开始，基于Java语言重新开发。提供了一个高性能的 RADIUS 处理引擎，同时提供了一个简洁易用的 WEB管理界面，可以轻松上手。

TOUGHRADIUS 的功能类似于 freeRADIUS，但它使用起来更简单，更易于扩展开发。

## 链接

- [网站首页](https://www.toughradius.net/)
- [TOUGHRADIUS WIKI 文档](https://github.com/talkincode/ToughRADIUS/wiki)
- [GUI 测试工具](https://github.com/jamiesun/RadiusTester)
- [命令行测试工具](https://github.com/talkincode/JRadiusTester)

## 快速开始

### 系统环境依赖

- 操作系统：支持跨平台部署 （Linux，Windows，MacOS等）
- java 版本: 1.8或更高
- 数据库服务器：MySQL/MariaDB

### 数据库初始化

> 数据库的安装配置请自行完成,首先确保你的数据库服务器已经运行

运行创建数据库脚本以及创建专用用户

    create database toughradius DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
    GRANT ALL ON toughradius.* TO raduser@'127.0.0.1' IDENTIFIED BY 'radpwd' WITH GRANT OPTION;FLUSH PRIVILEGES;

创建数据库表

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
    

插入测试数据

    INSERT INTO toughradius.tr_bras
    (identifier, name, ipaddr, vendor_id, portal_vendor,secret, coa_port,ac_port, auth_limit, acct_limit, STATUS, remark, create_time)
    VALUES ('radius-tester', 'radius-tester', '127.0.0.1', '14988',"cmccv1", 'secret', 3799,2000, 1000, 1000, NULL, '0', '2019-03-01 14:07:46');

    INSERT INTO toughradius.tr_subscribe
    (node_id,  subscriber, realname, password, domain, addr_pool, policy, is_online, active_num,
     bind_mac, bind_vlan, ip_addr, mac_addr, in_vlan, out_vlan, up_rate, down_rate, up_peak_rate, 
     down_peak_rate, up_rate_code,down_rate_code, status, remark, begin_time, expire_time, create_time, update_time)
    VALUES (0, 'test01', '', '888888',  null, null, null, null, 10, 0, 0, '', '', 0, 0, 10.000, 10.000, 100.000, 100.000,
            '10', '10', 'enabled', '', '2019-03-01 14:13:02', '2019-03-01 14:13:00', '2019-03-01 14:12:59', '2019-03-01 14:12:56');
            
### 运行主程序

    java -jar -Xms256M -Xmx1024M /opt/toughradius-latest.jar  --spring.profiles.active=prod
    
> 注意 jar 文件（toughradius-latest.jar）的路径

### Linux  systemd 服务配置

/opt/application-prod.properties

    # web访问端口
    server.port = 1816
    
    # 如果启用 https， 取消以下注释即可
    #server.security.require-ssl=true
    #server.ssl.key-store-type=PKCS12
    #server.ssl.key-store=classpath:toughradius.p12
    #server.ssl.key-store-password=toughstruct
    #server.ssl.key-alias=toughradius
    
    # 日志配置，可选 logback-prod.xml 或 logback-dev.xml， 日志目录为 /var/toughradius/logs
    logging.config=classpath:logback-prod.xml
    
    # 数据库配置
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

> 如果了解 spring systemd和配置原理，可以根据自己的实际需要进行修改

通过以下指令启动服务

    systemctl enable toughradius
    systemctl start toughradius
    
