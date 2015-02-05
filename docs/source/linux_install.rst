ToughRADIUS在linux下的安装配置
====================================

ToughRADIUS是基于Python及高性能异步网络框架Twisted开发，对linux系统完美支持，可以提供更高的性能和稳定性。

目前在Linux环境下，ToughRADIUS提供了自动化安装脚本，可以轻松的帮你完成安装过程。


已支持自动化安装的linux系统
------------------------------------

CentOS 6 , CentOS 7

脚本路径
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

install/centos-install

安装流程
------------------------------------

自动化安装过程在终端下执行,以CentOS 为例：

1. 下载脚本::

    $ curl https://raw.githubusercontent.com/talkincode/ToughRADIUS/master/install/centos-install > centos-install

    $ chmod +x centos-install

2. 执行安装::

    $ ./centos-install

在安装过程中会需要用户进行一些交互，如配置选项设置，是否安装本地mysql数据库。

执行完成以上两步可完成所有安装，然后就可以使用了。


分步骤安装
~~~~~~~~~~~~~~~~~~~~~~~~~

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


进程管理
------------------------------------

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