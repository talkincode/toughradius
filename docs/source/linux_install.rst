ToughRADIUS在linux下的安装配置
====================================

ToughRADIUS是基于Python及高性能异步网络框架Twisted开发，对linux系统完美支持，可以提供更高的性能和稳定性。

建议以CentOS7部署运行环境，ToughRADIUS需要Python2.7版本支持，CentOS7默认内置Python2.7版本，使得我们可以更方便的完成部署。

.. topic:: 注意

    以下安装的顺序不要乱，保证依赖不要出错。

一些系统依赖的模块安装
------------------------------------

安装以下系统模块，确保后续安装过成出现缺少模块的错误::

    $ yum install -y git gcc python-devel python-setuptools


MySQL数据库安装
--------------------------------

执行以下安装过程::

    $ rpm -ivh http://dev.mysql.com/get/mysql-community-release-el7-5.noarch.rpm
    
    $ yum install -y mysql-community-server mysql-community-devel 
    
初始化数据库::

    $ mysql_install_db
    
启动数据库::

    $ /usr/bin/mysqld_safe &
    
新增管理账号::

    $ echo "GRANT ALL ON *.* TO admin@'%' IDENTIFIED BY 'radius' WITH GRANT OPTION; FLUSH PRIVILEGES" | mysql
    


Python依赖模块安装
--------------------------------------

ToughRADIUS依赖一些特定的Python模块，我们也需要安装，才能保证ToughRADIUS的顺利运行

安装pip模块管理工具与supervisor进程管理工具::

    $ easy_install pip supervisor


安装ToughRADIUS依赖的Python模块::

    $ pip install -r https://raw.githubusercontent.com/talkincode/ToughRADIUS/master/requirements.txt


安装部署ToughRADIUS
------------------------------

使用git拉取ToughRADIUS版本::

    $ git clone https://github.com/talkincode/ToughRADIUS.git /opt/toughradius

或者::

    $ git clone https://coding.net/jamiesun/ToughRADIUS.git /opt/toughradius
    
建立配置文件::

    $ cp /opt/toughradius/config.json /etc/toughradius.json 
    
编辑/etc/toughradius.json文件，在配置文件中，可以指定几乎所有的配置参数。

databse部分是数据库的配置，修改配置文件数据库部分的主机，端口，用户名，数据库名，密码和实际相符合。

radiusd是Radius核心服务的配置，注意adminport是提供给web管理系统调用服务的端口，allows主要是web管理系统与自助服务系统的IP地址。

admin部分是web管理控制台配置，注意服务端口的配置，如果与系统其他应用冲突请修改。

customer是自助服务系统配置，注意服务端口的配置，如果与系统其他应用冲突请修改。

.. code-block:: javascript

    {
        "database": 
        {
            "dbtype":"mysql",
            "maxusage": 10, 
            "passwd": "radius",
            "charset": "utf8", 
            "db": "toughradius",
            "host": "192.168.59.103",
            "port": 3306,
            "user": "admin"
        },
        "radiusd":
        {
            "authport": 1812,
            "acctport": 1813,
            "adminport": 1815,
            "allows":"127.0.0.1",
            "dictfile": "./radiusd/dict/dictionary",
            "debug":1,
            "cache_timeout":600
        },
        "admin":
        {
            "httpport":1816,
            "debug":1
        },
        "customer":
        {
            "httpport":1817,
            "debug":1
        }    
    }
    
初始化ToughRADIUS数据库::

    $ cd /opt/toughradius && python createdb.py -c /etc/toughradius.json

按提示完成操作::

    starting create and init database...
    drop and create database ?[n]y
    init database ?[n]y


启动ToughRADIUS服务
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

启动RADIUS核心认证计费授权服务::

    $ cd /opt/toughradius \
        && python radiusd/server.py -c /etc/toughradius.json 

以守护进程模式运行::

    $ cd /opt/toughradius \
        && nohup python radiusd/server.py -c /etc/toughradius.json  &

启动Web管理控制台系统::

    $ cd /opt/toughradius/console \
        && python admin.py -c /etc/toughradius.json

以守护进程模式运行::

    $ cd /opt/toughradius/console \
        && nohup python admin.py -c /etc/toughradius.json &


使用supervisor进程管理工具来部署ToughRADIUS
-------------------------------------------

supervisor是一个进程管理工具，本身也是python的一个模块

建立supervisor配置文件/etc/supervisord.conf::

    $ vi /etc/supervisord.conf

配置文件内容::

    [unix_http_server]
    file=/tmp/supervisor.sock   ; (the path to the socket file)


    [inet_http_server]         ; inet (TCP) server disabled by default
    port=127.0.0.1:9001        ; (ip_address:port specifier, *:port for all iface)

    [supervisord]
    logfile=/var/toughradius/log/supervisord.log ; (main log file;default $CWD/supervisord.log)
    logfile_maxbytes=50MB        ; (max main logfile bytes b4 rotation;default 50MB)
    logfile_backups=10           ; (num of main logfile rotation backups;default 10)
    loglevel=info                ; (log level;default info; others: debug,warn,trace)
    pidfile=/tmp/supervisord.pid ; (supervisord pidfile;default supervisord.pid)
    nodaemon=false               ; (start in foreground if true;default false)
    minfds=1024                  ; (min. avail startup file descriptors;default 1024)
    minprocs=200                 ; (min. avail process descriptors;default 200)


    [rpcinterface:supervisor]
    supervisor.rpcinterface_factory = supervisor.rpcinterface:make_main_rpcinterface


    [supervisorctl]
    serverurl=http://127.0.0.1:9001 ; use an http:// url to specify an inet socket

    [program:radiusd]
    command=/usr/bin/python radiusd/server.py -c /etc/toughradius.json
    process_name=%(program_name)s
    numprocs=1
    directory=/opt/toughradius
    autostart=true
    autorestart=true
    user=root
    redirect_stderr=true
    stdout_logfile=/var/log/radiusd.log

    [program:rad_console]
    command=/usr/bin/python admin.py -c /etc/toughradius.json
    process_name=%(program_name)s
    numprocs=1
    directory=/opt/toughradius/console
    autostart=true
    autorestart=true
    user=root
    redirect_stderr=true
    stdout_logfile=/var/log/rad_console.log

    [program:rad_customer]
    command=/usr/bin/python customer.py -c /etc/toughradius.json
    process_name=%(program_name)s
    numprocs=1
    directory=/opt/toughradius/console
    autostart=true
    autorestart=true
    user=root
    redirect_stderr=true
    stdout_logfile=/var/log/rad_customer.log


*启动(守护进程模式)*::

    $ supervisord -c /etc/supervisord.conf 

*查看状态*::

    $ supervisorctl status
    rad_customer                     RUNNING   pid 32132, uptime 3:35:24
    rad_console                      RUNNING   pid 32133, uptime 3:35:25
    radiusd                          RUNNING   pid 32130, uptime 3:35:28

*其他控制指令*::

    $ supervisorctl start all
    $ supervisorctl stop all
    $ supervisorctl restart all

    # 指定具体的进程

    $ supervisorctl start radiusd
    $ supervisorctl stop radiusd

    # 如果修改了/etc/supervisord.conf 
    $ supervisorctl reload








