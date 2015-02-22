# ToughRADIUS 

ToughRADIUS是一个开源的Radius服务软件，基于BSD许可协议发布。

ToughRADIUS支持标准RADIUS协议，提供完整的AAA实现。支持灵活的策略管理，支持各种主流接入设备并轻松扩展，具备丰富的计费策略支持。

ToughRADIUS支持使用Oracle, MySQL, PostgreSQL, MSSQL等主流数据库存储用户数据，并支持数据缓存，极大的提高了性能。
ToughRADIUS支持Windows，Linux，BSD跨平台部署，部署使用简单。

ToughRADIUS提供了RADIUS核心服务引擎与Web管理控制台,用户自助服务三个子系统，核心服务引擎提供高性能的认证计费服务，Web管理控制台提供了界面友好，功能完善的管理功能。用户自助服务系统提供了一个面向终端用户的网上服务渠道。

[ToughRADIUS网站：http://www.toughradius.net] (http://www.toughradius.net)

[ToughRADIUS文档: http://docs.toughradius.net/build/html/] (http://docs.toughradius.net/build/html/)

## 安装

1. 安装系统依赖（centos6/7）

    yum update -y
    
    # centos 6
    yum install -y  mysql-devel python-devel python-setuptools MySQL-python
    
    #centos7
    yum install -y  mariadb-devel python-devel python-setuptools MySQL-python
    
    easy_install pip supervisor
    
   
2. 安装toughradius,安装完成后，toughctl，toughrad命令可用。

    pip install toughradius
    

3. 创建配置文件,请确保你的mysql服务器已经安装运行，根据提示配置正确的数据库连接信息。

    toughctl --config
    
    [INFO] - set config...
    [INPUT] - set your config file path,[ /etc/radiusd.conf ]
    [INFO] - set default option
    [INPUT] - set debug [0/1] [0]:
    [INPUT] - time zone [ CST-8 ]:
    [INFO] - set database option
    [INPUT] - database type [mysql]:
    [INPUT] - database host [127.0.0.1]:
    [INPUT] - database port [3306]:
    [INPUT] - database dbname [toughradius]:
    [INPUT] - database user [root]:
    [INPUT] - database passwd []:
    [INPUT] - db pool maxusage [30]:
    [INFO] - set radiusd option
    [INPUT] - radiusd authport [1812]:
    [INPUT] - radiusd acctport [1813]:
    [INPUT] - radiusd adminport [1815]:
    [INPUT] - radiusd cache_timeout (second) [600]:
    [INPUT] - log file [ logs/radiusd.log ]:/var/log/radiusd.log
    [INFO] - set mysql backup ftpserver option
    [INPUT] - backup ftphost [127.0.0.1]:
    [INPUT] - backup ftpport [21]:
    [INPUT] - backup ftpuser [ftpuser]:
    [INPUT] - backup ftppwd [ftppwd]:
    [INFO] - set admin option
    [INPUT] - admin http port [1816]:
    [INPUT] - log file [ logs/admin.log ]:/var/log/admin.log
    [INFO] - set customer option
    [INPUT] - customer http port [1817]:
    [INPUT] - log file [ logs/customer.log ]:/var/log/customer.log
    [SUCC] - config save to /etc/radiusd.conf
    
4. 初始化数据库
    
    #还未创建数据库，使用参数 initdb 1 或 initdb 2
    toughctl --initdb 1
    
    #已创建数据库，使用参数 initdb 3
    toughctl --initdb 3
    
5. 运行服务

    #radius认证计费服务
    toughctl --radiusd
    
    #radius管理控制台服务
    toughctl --admin
    
    #radius用户自助服务
    toughctl --customer
    

6. 以守护服务模式运行（supervisor）

    #加入supervisord配置文件
    toughctl ---echo_supervisord_cnf > /etc/supervisord.conf
    
    #启动supervisord进程管理服务
    supervisord -c /etc/supervisord.conf
    
    
## web管理控制台的使用

当安装部署完成后可使用浏览器进入管理控制台进行操作。

默认地址与端口：http://serverip:1816
    
默认管理员与密码：admin/root

## 自助服务系统的使用

自助服务系统运行于一个独立的进程。

默认地址与端口:http://serverip:1817
