# ToughRADIUS  [![Build Status](https://travis-ci.org/talkincode/ToughRADIUS.svg?branch=master)](https://travis-ci.org/talkincode/ToughRADIUS)

ToughRADIUS是一个开源，免费，易用的Radius服务软件。

ToughRADIUS支持标准RADIUS协议，提供完整的AAA实现。支持灵活的策略管理，支持各种主流接入设备并轻松扩展，完美对接RouterOS，支持包月，时长计费。

ToughRADIUS支持MySQL存储用户数据，支持数据缓存，极大的提高了性能。

ToughRADIUS支持Windows，Linux跨平台部署，部署使用简单。

提供了RADIUS核心服务引擎与Web管理控制台,用户自助服务三个子系统，核心服务引擎提供高性能的认证计费服务，Web管理控制台提供了界面友好，功能完善的管理功能。用户自助服务系统提供了一个面向终端用户的网上服务渠道。

[ToughRADIUS网站：http://www.toughradius.net] (http://www.toughradius.net)

[ToughRADIUS文档: http://docs.toughradius.net/build/html/] (http://docs.toughradius.net/build/html/)

## Linux下使用脚本自动安装

目前在Linux环境下，ToughRADIUS提供了自动化安装脚本，可以轻松的帮你完成安装过程。

已支持自动化安装的linux系统

CentOS 6 , CentOS 7

脚本路径

    install/centos-install

    
### 安装过程

自动化安装过程在终端下执行,以CentOS为例：

下载脚本::

    $ curl https://raw.githubusercontent.com/talkincode/ToughRADIUS/master/install/centos-install > centos-install

    $ chmod +x centos-install

执行安装::

    $ ./centos-install

在安装过程中会需要用户进行一些交互，如配置选项设置，是否安装本地mysql数据库。

执行完成以上两步可完成所有安装，然后就可以使用了。


#### 分步安装

同时该脚本也提供了分步骤安装的支持。

安装系统必要的依赖库请执行::

    $ ./centos-install depend
    
安装ToughRADIUS请执行::

    $ ./centos-install radius

安装mysql(可选)请执行::

    $ ./centos-install mysql

定义ToughRADIUS配置执行::
    
    # 如果你选择不在本机安装mysql数据库，应该注意配置你的远程数据库参数

    $ ./centos-install config

创建ToughRADIUS数据库请执行::

    $ ./centos-install initdb
    
完成以上所有后快速启动ToughRADIUS::

    # 在start之前请确认你的配置无误

    $ ./centos-install start 

### 进程管理

通过以上步骤安装完成后，会提供一个进程管理工具 toughrad

启动ToughRADIUS进程::

    $ toughrad start

停止ToughRADIUS进程::

    $ toughrad stop

重启ToughRADIUS进程::

    $ toughrad restart
    
升级ToughRADIUS到最新版本::

    $ toughrad upgrade    

启动mysql数据库进程::

    $ toughrad startdb

停止mysql数据库进程::

    $ toughrad stopdb
    
备份ToughRADIUS主数据库,备份路径在/var/toughradius/databak,若要上传至ftp，请配置/var/toughradius/radiusd.json文件中的备份选项::

    $ toughrad backupdb

跟踪数据库日志::

    $ toughrad tracedb
    
跟踪radius日志::

    $ toughrad tracerad
    
跟踪管理控制台日志::

    $ toughrad traceadmin
    
跟踪自助服务控制台日志::

    $ toughrad tracecustomer    

## 使用Docker镜像 

在centos7下部署::

    $ yum install docker

    $ service docker start

    $ mkdir /var/toughradius

    $ docker run -d -v /var/toughradius:/var/toughradius \
      -p 3306:3306 -p 1812:1812/udp -p 1813:1813/udp \
      -p 1815:1815 -p 1816:1816 -p 1817:1817\
      --name toughradius talkincode/centos7-toughradius 

以上指令会在centos7中安装docker工具，并自动下载toughradius镜像以守护进程模式运行。

运行 docker logs toughradius 即可看到运行日志输出。


## web管理控制台的使用

当安装部署完成后可使用浏览器进入管理控制台进行操作。

默认地址与端口：http://serverip:1816
    
默认管理员与密码：admin/root

## 自助服务系统的使用

自助服务系统运行于一个独立的进程。

默认地址与端口:http://serverip:1817
