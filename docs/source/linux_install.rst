ToughRADIUS在linux下的安装配置
====================================

ToughRADIUS是基于Python及高性能异步网络框架Twisted开发，对linux系统完美支持，可以提供更高的性能和稳定性。

以下安装环境以CentOS5.8_x64为例，

一些系统依赖的模块安装
------------------------------------

yum install -y zlib zlib-devel bzip2 bzip2-devel

安装以上系统模块，确保后续安装过成出现缺少模块的错误。


Python环境安装配置(源码编译安装)
--------------------------------------

ToughRADIUS要求Python2.7版本（不支持python3），在CentOS5.8_x64中内置的Python版本过低（2.4.3），因此，我们需要自己编译Python版本。

从Python官网下载Python源码包，按下面指示完成安装::

    cd /usr/local/src 

    wget https://www.python.org/ftp/python/2.7.6/Python-2.7.6.tgz

    tar zxvf Python-2.7.6.tgz

    cd Python-2.7.6

    ./configure

    make && make install

输入指令::

    /usr/local/bin/python2.7 -V 

输出::

    Python 2.7.6

说明python安装成功

.. topic:: 注意

    非必要情况不要覆盖系统本身的python2.4.3版本，比如一些服务yum，ibus必须依赖系统的python版本。

MySQL数据库安装
--------------------------------

::

    yum -y install mysql-server mysql mysql-devel

开机启动配置::

    $ sudo /sbin/chkconfig --add mysqld
    $ sudo /sbin/chkconfig mysqld on   
    $ sudo /sbin/service mysqld start

设置密码::
    
    $ sudo mysqladmin -u root password 'mypassword'


安装ToughRADIUS依赖Python模块
-----------------------------------------

下载模块依赖描述文件::

    wget https://raw.githubusercontent.com/talkincode/ToughRADIUS/master/requirements.txt

requirements.txt中指定了ToughRADIUS依赖的python模块::

    DBUtils==1.1
    Mako==0.9.0
    Beaker==1.6.4
    MarkupSafe==0.18
    MySQL-python==1.2.5
    PyYAML==3.10
    SQLAlchemy==0.9.8
    Twisted==14.0.2
    autobahn==0.9.3-3
    bottle==0.12.7
    six==1.8.0
    tablib==0.10.0
    zope.interface==4.1.1
    pycrypto==2.6.1

要完成安装扩展模块，必须保证主机处于联网状态::

    # 安装setuptools, setuptools是python的包管理工具

    wget https://pypi.python.org/packages/source/s/setuptools/setuptools-3.3.zip

    unzip setuptools-3.3.zip 

    cd setuptools-3.3

    python2.7 setup.py install 

    # 安装模块 pip, 另一个更好的包管理工具
    
    easy_install pip 

    # 批量安装 依赖模块

    pip install -r requirements.txt

    # 可以尝试增加 -i 参数指定豆瓣服务器镜像源来加速模块下载速度

    pip install -r requirements.txt -i http://pypi.douban.com/simple


安装部署ToughRADIUS
------------------------------

下载ToughRADIUS发布版本(版本以实际发布为准)::

    cd /opt 
    wget https://github.com/talkincode/ToughRADIUS/archive/v0.1.zip -O toughradius.zip
    # 或者下载最新版本
    wget https://github.com/talkincode/ToughRADIUS/archive/master.zip -O toughradius.zip

如果github速度慢，可选择国内coding源::

    cd /opt 
    wget https://coding.net/u/jamiesun/p/ToughRADIUS/git/archive/v0.1 -O toughradius.zip
    # 或者下载最新版本
    wget https://coding.net/u/jamiesun/p/ToughRADIUS/git/archive/master -O toughradius.zip

解压缩版本::

    unzip toughradius.zip && cd toughradius


应用配置说明
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

在config.json文件中，可以指定几乎所有的配置参数，同时允许自定义命令行参数，命令行参数会覆盖配置文件的定义。

修改配置文件mysql部分的主机，用户名，数据库名，密码和实际相符合。

.. code-block:: javascript

    {
        "mysql": 
        {
            "maxusage": 10, 
            "passwd": "root",
            "charset": "utf8", 
            "db": "toughradius",
            "host": "10.211.55.2",
            "user": "root"
        },
        "radiusd":
        {
            "authport": 1812,
            "acctport": 1813,
            "adminport": 1815,
            "dictfile": "./radiusd/dict/dictionary",
            "debug":1,
            "cache_timeout":600
        },
        "console":
        {
            "httpport":1816,
            "radaddr":"127.0.0.1",
            "adminport":1815,
            "debug":1
        }
    }

.. topic:: 注意

    在实际环境中radaddr必须填写真实地radiusd服务IP地址或主机名，不要使用本地地址。

    admin端口是radiusd的管理监听端口，在console中会通过该端口调用一些管理服务，比如实时查询跟踪用户消息等。


创建ToughRADIUS数据库
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

可以使用sql脚本创建::

    #登录mysql client
    mysql -u root -p

    create database toughradius DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;

    use toughradius;

    #执行建表脚本,注意(sql脚本以实际发布的版本脚本文件为准)
    source /opt/ToughRADIUS/toughradius.sql；

    #完成退出
    quit;

也可以使用create.py脚本来创建，运行脚本::

    python2.7 createdb.py -c config.json

按提示完成操作::

    starting create and init database...

    drop and create database ?[n]y

    init database ?[n]y

    init testdata ?[n]n

启动ToughRADIUS服务
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

启动RADIUS核心认证计费授权服务::

    python2.7 radiusd/server.py -c config.json

以守护进城模式运行::

    nohup python2.7 radiusd/server.py -c config.json &

启动Web管理控制台系统::
    
    cd console

    python2.7 admin.py -c ../config.json

以守护进城模式运行::

    nohup python2.7 admin.py -c ../config.json &


使用supervisor进程管理工具来部署ToughRADIUS
-------------------------------------------

supervisor是一个进程管理工具，本身也是python的一个模块

*安装*::

    pip install supervisor

*配置*::

    # 安装完supervisor就有了这个工具，生成配置文件
    echo_supervisord_conf > /etc/supervisord.conf

    # 在 /etc/supervisord.conf  末尾加入内容 

    ... ...

    [program:radiusd]
    command=/usr/local/bin/python2.7 radiusd/server.py -c config.json
    process_name=%(program_name)s
    numprocs=1
    directory=/opt/ToughRADIUS
    autostart=true
    autorestart=true
    user=root
    redirect_stderr=true
    stdout_logfile=/var/log/radiusd.log

    [program:rad_console]
    command=/usr/local/bin/python2.7 admin.py -c ../config.json
    process_name=%(program_name)s
    numprocs=1
    directory=/opt/ToughRADIUS/console
    autostart=true
    autorestart=true
    user=root
    redirect_stderr=true
    stdout_logfile=/var/log/rad_console.log    

*启动*::

    supervisord -c /etc/supervisord.conf 

*查看状态*::

    [root@server ~]# supervisorctl status
    rad_console                      RUNNING   pid 32133, uptime 3:35:25
    radiusd                          RUNNING   pid 32130, uptime 3:35:28

*其他控制指令*::

    supervisorctl start all
    supervisorctl stop all
    supervisorctl restart all

    # 指定具体的进程

    supervisorctl start radiusd
    supervisorctl stop radiusd

    # 如果修改了/etc/supervisord.conf 
    supervisorctl reload








