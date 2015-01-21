ToughRADIUS在windows下的安装配置
====================================

ToughRADIUS为windows提供了一个快速部署的模式，帮助使用者快速部署ToughRADIUS服务。


最新版本下载
--------------------------------

从以下链接可以下载最新的ToughRADIUS版本：

`github.com mirror <https://github.com/talkincode/ToughRADIUS/archive/master.zip>`_

`coding.net mirror <https://coding.net/u/jamiesun/p/ToughRADIUS/git/archive/master>`_

`oschina.net mirror <https://git.oschina.net/jamiesun/ToughRADIUS/repository/archive?ref=master>`_


数据库安装配置
--------------------------------

ToughRADIUS主要采用MySQL(5.0以上版本)存储数据，在部署ToughRADIUS之前请自行安装MySQL（安装步骤请参考MySQL相关文档）,安装MySQL后确保MySQL为运行状态。

+ 修改配置文件 config.json中的mysql选项,请修改数据库地址用户名密码等选项与实际相符。

.. code-block:: javascript

    {
        "mysql": 
        {
            "maxusage": 10, 
            "passwd": "root",
            "charset": "utf8", 
            "db": "toughradius",
            "host": "127.0.0.1",
            "user": "root"
        }
    }

+ 运行createdb.bat创建数据库表，ToughRADIUS采用脚本工具自动创建数据库，无需SQL脚本。

在windows环境下，双击createdb.bat即可进行数据库创建过程。

.. code-block:: bash

    createdb.bat

    #按提示进行操作

    Z:\github\ToughRADIUS>toughrad.exe createdb.py  -c config.json  || pause

    starting create and init database...

    drop and create database ?[n]y

    init database ?[n]y

    init testdata ?[n]n

.. topic:: 注意

    运行脚本会尝试删除原有数据库并重建，如果非首次安装，建议备份数据，init testdata是创建测试数据选项，一般不需要。


应用配置说明
-------------------------------

在config.json文件中，可以指定几乎所有的配置参数，同时允许自定义命令行参数，命令行参数会覆盖配置文件的定义。

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


运行radiusd服务
--------------------------------

radiusd提供提供了RADIUS核心认证计费授权服务，在windows环境下，双击radiusd.bat脚本即可运行。

radiusd.bat内容

.. code-block:: bash

    toughrad.exe radiusd/server.py -c config.json || pause   


示例：

.. code-block:: bash

    radiusd.bat

    Z:\github\ToughRADIUS>toughrad.exe radiusd/server.py -c config.json  -dict radiu
    sd/dict/dictionary    || pause

    ['radiusd/server.py', '-c', 'config.json', '-dict', 'radiusd/dict/dictionary']

    logging to file logs/radiusd.log


默认情况下，日志会打印到logs/radiusd.log文件里，在debug模式下将会打印系统更详细的日志，并会在控制台实时输出。

运行console服务
--------------------------------

console是Web管理控制台系统，在windows环境下，双击console.bat脚本即可运行。

console.bat脚本内容

.. code-block:: bash

    cd console && ..\toughrad.exe admin.py -c ../config.json || pause

示例：

.. code-block:: bash

    console.bat

    Z:\github\ToughRADIUS>cd console   && ..\toughrad.exe admin.py -c ../config.json || pause
    Z:\github\ToughRADIUS\console
    Z:\github\ToughRADIUS\console
    ToughRADIUS Console Server Starting up...
    Listening on http://0.0.0.0:1816/
    Hit Ctrl-C to quit.

当启动web控制台服务后，就可以通过浏览器访问管理界面了，在浏览器地址栏输入：http://127.0.0.1:1816,默认的管理员密码为admin/root

登陆界面：

.. image:: ./_static/images/toughradius_login.jpg





