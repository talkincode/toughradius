# README

[![](https://badge.imagelayers.io/talkincode/toughradius:v2.svg)](https://imagelayers.io/?images=talkincode/toughradius:v2 'Get your own badge on imagelayers.io')

# TOUGHRADIUS 简介

TOUGHRADIUS是一个开源的Radius服务软件，采用于AGPL许可协议发布。

TOUGHRADIUS支持标准RADIUS协议，提供完整的AAA实现。支持灵活的策略管理，支持各种主流接入设备并轻松扩展，具备丰富的计费策略支持。

TOUGHRADIUS支持使用Oracle, MySQL, PostgreSQL, MSSQL等主流数据库存储用户数据，并支持数据缓存，极大的提高了性能。

TOUGHRADIUS支持Windows，Linux，BSD跨平台部署，部署使用简单。

TOUGHRADIUS提供了RADIUS核心服务引擎与Web管理控制台,用户自助服务三个子系统，核心服务引擎提供高性能的认证计费服务，Web管理控制台提供了界面友好，功能完善的管理功能。用户自助服务系统提供了一个面向终端用户的网上服务渠道。

TOUGHRADIUS网站：http://www.toughradius.net


# TOUGHRADIUS 文档


http://docs.toughradius.net

# 快速指南

## 服务器安装配置

### Docker 安装

*CentOS 7*

    $ yum install docker
    $ service docker start

*其他linux系统*

    $ curl -sSL https://get.daocloud.io/docker | sh

### 安装 docker-compose

    $ easy_install docker-compose

或者

    curl -L https://get.daocloud.io/docker/compose/releases/download/1.5.2/docker-compose-`uname -s`-`uname -m` > /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose

*Windows*

> 请下载[Windows安装文件](https://github.com/boot2docker/windows-installer/releases/download/v1.8.0/docker-install.exe)：[https://get.daocloud.io/toolbox/windows](https://get.daocloud.io/toolbox/windows)

### 部署配置描述文件

配置内容根据实际修改，如果想链接已经存在的数据库，可将 raddb的部分删除，并修改数据库连接部分配置。

    $ mkdir -p /opt/radius
    
    $ curl https://github.com/talkincode/ToughRADIUS/raw/master/docker-compose.simple.yml > /opt/radius/docker-compose.yml
    
    $ vi /opt/radius/docker-compose.yml
    
    raddb:
    image: "index.alauda.cn/toughstruct/mysql:512M"
    privileged: true
    expose:
        - "3306"
    environment:
        - MYSQL_USER=raduser
        - MYSQL_PASSWORD=radpwd
        - MYSQL_DATABASE=radiusd
        - MYSQL_ROOT_PASSWORD=radroot
    restart: always
    volumes:
        - /home/toughrun/trmysql:/var/lib/mysql
    
    radius:
    images: "index.alauda.cn/toughstruct/toughradius:v2"
    ports:
        - "1816:1816"
        - "1812:1812/udp"
        - "1813:1813/udp"
    links:
        - raddb:raddb
    environment:
        - DB_TYPE=mysql
        - DB_URL=mysql://raduser:radpwd@raddb:3306/radiusd?charset=utf8
    restart: always
    volumes:
        - /home/toughrun/toughradius:/var/toughradius

### 部署实例

    $ cd /opt/radius
    
    $ docker-compose up -d

使用 docker-compose ps 可以查看实例的运行状态，如果看到都是up，说明实例已经正确运行。

访问 http://server:1816  进入管理系统，默认用户名密码是 admin/root

### 部署更新

如果修改了 /opt/radius/docker-compose.yml 文件，只需要重新运行一次 docker-compose up -d 即可


# TOUGHRADIUS 商业授权

[TOUGHRADIUS 商业授权](#) (https://github.com/talkincode/ToughRADIUS/blob/master/Commerical-license.rst)