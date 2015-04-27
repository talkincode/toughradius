ToughRADIUS  Windows Ver.
====================================
Power BY. LJTECH

ToughRADIUS是一个开源的Radius服务软件，基于BSD许可协议发布。

ToughRADIUS支持标准RADIUS协议，提供完整的AAA实现。支持灵活的策略管理，支持各种主流接入设备并轻松扩展，具备丰富的计费策略支持。

ToughRADIUS支持使用Oracle, MySQL, PostgreSQL, MSSQL等主流数据库存储用户数据，并支持数据缓存，极大的提高了性能。
ToughRADIUS支持Windows，Linux，BSD跨平台部署，部署使用简单。

ToughRADIUS提供了RADIUS核心服务引擎与Web管理控制台,用户自助服务三个子系统，核心服务引擎提供高性能的认证计费服务，Web管理控制台提供了界面友好，功能完善的管理功能。用户自助服务系统提供了一个面向终端用户的网上服务渠道。

ToughRADIUS网站：http://www.toughradius.net

ToughRADIUS文档: http://docs.toughradius.net/build/html/



Linux环境快速安装
====================================


安装系统依赖(centos6/7)
--------------------------------------

::

    $ yum update -y  && yum install -y  python-devel python-setuptools 
    
    $ easy_install pip
    
    
    
安装toughradius
----------------------------------------

安装完成后，toughctl命令可用。

::

    $ pip install toughradius
    

系统配置
----------------------------------------

::

    $ toughctl --echo_radiusd_cnf > /etc/radiusd.conf
    
配置文件内容::

    [DEFAULT]
    debug = 0
    tz = CST-8
    secret = %s

    [database]
    dbtype = sqlite
    dburl = sqlite:////tmp/toughradius.sqlite3
    echo = false

    [radiusd]
    acctport = 1813
    adminport = 1815
    authport = 1812
    cache_timeout = 600
    logfile = /var/log/radiusd.log

    [admin]
    port = 1816
    logfile = /var/log/admin.log

    [customer]
    port = 1817
    logfile = /var/log/customer.log


初始化数据库
----------------------------------------

注意此操作会重建所有数据库表，请注意备份重要数据。

::

    $ toughctl --initdb 


运行服务
----------------------------------------

::

    $ toughctl --standalone
    

以守护进程模式运行
----------------------------------------

当启动standalone模式时，只会启动一个进程

::

    # 启动
    
    $ toughctl --start standalone 
    
    # 停止
    
    $ toughctl --stop standalone 
     
    # 设置开机启动
    
    $ echo "toughctl --start standalone" >> /etc/rc.local
    
    
web管理控制台的使用
================================

当安装部署完成后可使用浏览器进入管理控制台进行操作。

默认地址与端口：http://serverip:1816 
 
默认管理员与密码：admin/root


自助服务系统的使用
================================

自助服务系统运行于一个独立的进程。

默认地址与端口:http://serverip:1817
