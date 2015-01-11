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

+ 修改配置文件 config.json,请修改数据库地址用户名密码等选项与实际相符。

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

    Z:\github\ToughRADIUS>toughrad.exe createdb.py    || pause

    starting create and init database...

    drop and create database ?[n]y

    init database ?[n]y

    init testdata ?[n]n

.. topic:: 注意

    运行脚本会尝试删除原有数据库并重建，如果非首次安装，建议备份数据，init testdata是创建测试数据选项，一般不需要。


运行radiusd服务
--------------------------------

radiusd提供提供了RADIUS核心认证计费授权服务，在windows环境下，双击radiusd.bat脚本即可运行。

radiusd.bat内容

.. code-block:: bash

    toughrad.exe radiusd/server.py -c config.json  -dict radiusd/dict/dictionary || pause   

你可以新建一个debug的脚本，加上 -d 或者 --debug 参数即可。

.. code-block:: bash

    toughrad.exe radiusd/server.py -c config.json  -dict radiusd/dict/dictionary -d || pause

你可以通过参数指定端口

.. code-block:: bash

    toughrad.exe radiusd/server.py -auth 1812 -acct 1813 -admin 1815 -c config.json  -dict radiusd/dict/dictionary -d || pause

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

你可以新建一个debug的脚本，加上 -d 或者 --debug 参数即可。也可以指定端口运行(默认的http端口是1816)。

.. code-block:: bash

    cd console && ..\toughrad.exe admin.py -http 8080 -admin 1815 -c ../config.json || pause

示例：

.. code-block:: bash

    console.bat

    Z:\github\ToughRADIUS>cd console   && ..\toughrad.exe admin.py -c ../config.json || pause
    Z:\github\ToughRADIUS\console
    Z:\github\ToughRADIUS\console
    ToughRADIUS Console Server Starting up...
    Listening on http://0.0.0.0:1816/
    Hit Ctrl-C to quit.

当启动web控制台服务后，就可以通过浏览器访问管理界面了，在浏览器地址栏输入：http://127.0.0.1:1816


.. topic:: 注意

    admin端口是radiusd的管理监听端口，在console中会通过该端口调用一些管理服务，比如实时查询跟踪用户消息等。





