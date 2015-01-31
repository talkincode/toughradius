ToughRADIUS在linux下的安装配置
====================================

ToughRADIUS是基于Python及高性能异步网络框架Twisted开发，对linux系统完美支持，可以提供更高的性能和稳定性。

目前在Linux环境下，ToughRADIUS提供了自动化安装脚本，可以轻松的帮你完成安装过程。

.. topic:: 注意

    1, 自动安装脚本会安装mysql数据库

已支持自动化安装的linux系统
------------------------------------

CentOS 6 , CentOS 7


脚本路径
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

install/centos6-install.sh
install/centos7-install.sh


安装流程
------------------------------------

自动化安装过程在终端下执行,以CentOS 7为例：

1. 下载脚本

::
    $ curl https://raw.githubusercontent.com/talkincode/ToughRADIUS/master/install/centos7-install.sh > centos7-install.sh

2. 执行安装

::
    $ sh centos7-install.sh all

执行完成以上两步可完成所有安装并运行ToughRADIUS服务，然后就可以使用了。


分步骤安装
~~~~~~~~~~~~~~~~~~~~~~~~~

同时该脚本也提供了分步骤安装的支持。

安装系统必要的依赖库请执行::

    $ sh centos7-install.sh depend

安装mysql请执行::

    $ sh centos7-install.sh mysql5

安装ToughRADIUS请执行::

    $ sh centos7-install.sh radius

执行数据库初始化并启动ToughRADIUS请执行::

    $ sh centos7-install.sh setup

停止服务并卸载安装数据请执行::

    $ sh centos7-install.sh unsetup

