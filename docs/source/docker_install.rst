使用Docker镜像部署ToughRADIUS
=======================================

Docker 是一个开源项目，诞生于 2013 年初，最初是 dotCloud 公司内部的一个业余项目。它基于 Google 公司推出的 Go 语言实现。 项目后来加入了 Linux 基金会，遵从了 Apache 2.0 协议，项目代码在 GitHub 上进行维护。

Docker 自开源后受到广泛的关注和讨论，以至于 dotCloud 公司后来都改名为 Docker Inc。Redhat 已经在其 RHEL6.5 中集中支持 Docker；Google 也在其 PaaS 产品中广泛应用。

Docker 项目的目标是实现轻量级的操作系统虚拟化解决方案。 Docker 的基础是 Linux 容器（LXC）等技术。

在 LXC 的基础上 Docker 进行了进一步的封装，让用户不需要去关心容器的管理，使得操作更为简便。用户操作 Docker 的容器就像操作一个快速轻量级的虚拟机一样简单。


Docker 安装
-------------------------------

关于Docker安装的更多详细内容请见：http://docs.docker.com/installation/

CentOS6
~~~~~~~~~~~~~~~~~~~~~~~~~

::

    $ sudo yum install http://mirrors.yun-idc.com/epel/6/i386/epel-release-6-8.noarch.rpm
   
    $ sudo yum install docker-io

    $ sudo service docker start


CentOS7
~~~~~~~~~~~~~~~~~~~~~~~~~

::

    $ sudo yum install docker

    $ sudo service docker start


Ubuntu
~~~~~~~~~~~~~~~~~~~~~~~~~

安装最新版本的Ubuntu包（可能不是最新的docker版本包）::

    $ sudo apt-get update
    $ sudo apt-get install docker.io
    $ sudo ln -sf /usr/bin/docker.io /usr/local/bin/docker
    $ sudo sed -i '$acomplete -F _docker docker' /etc/bash_completion.d/docker.io

Windows
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

请下载Windows安装文件：https://github.com/boot2docker/windows-installer/releases/download/v1.4.1/docker-install.exe

运行ToughRADIUS
------------------------------------

关于容器的概念，你可以简单地理解为轻量的虚拟机。

创建并运行容器
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

通过Docker部署ToughRADIUS，需要创建一个容器，以后的更新可以直接在此容器上进行。

首先创建一个本地目录 /var/toughradius,Docker会利用此目录来创建mysql的数据库文件，以及配置文件，以后备份此目录即可。

::

    $ mkdir /var/toughradius 

    $ docker run -d -P -v /var/toughradius:/var/toughradius \
      -p 3306:3306 -p 1812:1812/udp -p 1813:1813/udp \
      -p 1815:1815 -p 1816:1816 \
      --name toughradius talkincode/centos7-toughradius

以上指令自动下载toughradius镜像,创建名称为toughradius的容器，以守护进程模式运行，容器只需创建一次，以上命令只需首次运行即可。

容器将本身端口与主机一一映射，如果有端口冲突请自行修改，格式 -p 主机端口:容器端口

运行 docker ps -a 可以看到容器进程信息

运行 docker logs toughradius 查看容器日志输出

如果你看到以下日志内容，说明运行成功了::

    150124 16:26:58 mysqld_safe Logging to '/var/toughradius/log/mysqld.log'.
    150124 16:26:58 mysqld_safe Starting mysqld daemon with databases from /var/toughradius/mysql
    starting create and init database...
    150124 16:27:05 mysqld_safe mysqld from pid file /var/run/mysqld/mysqld.pid ended
    starting mysqd...
    150124 16:27:07 mysqld_safe Logging to '/var/toughradius/log/mysqld.log'.
    150124 16:27:07 mysqld_safe Starting mysqld daemon with databases from /var/toughradius/mysql
    starting supervisord...
    2015-01-24 16:27:15,055 CRIT Supervisor running as root (no user in config file)
    2015-01-24 16:27:15,072 INFO RPC interface 'supervisor' initialized
    2015-01-24 16:27:15,073 CRIT Server 'unix_http_server' running without any HTTP authentication checking
    2015-01-24 16:27:15,073 INFO supervisord started with pid 420
    2015-01-24 16:27:16,076 INFO spawned: 'rad_console' with pid 423
    2015-01-24 16:27:16,078 INFO spawned: 'radiusd' with pid 424
    2015-01-24 16:27:17,136 INFO success: rad_console entered RUNNING state, process has stayed up for > than 1 seconds (startsecs)
    2015-01-24 16:27:17,136 INFO success: radiusd entered RUNNING state, process has stayed up for > than 1 seconds (startsecs)

打开浏览器访问 http://serverip:1816,可以进入web管理登陆界面了。


启动，停止，重启容器
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

    $ docker start toughradius

    $ docker stop toughradius

    $ docker restart toughradius


ToughRADIUS版本更新
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

当ToughRADIUS版本更新时，不需要重新创建容器，只需要执行简单地更新指令即可::

    $ docker exec toughradius sh /opt/upgrade.sh

    # 输出以下内容说明更新成功

    starting upgrade...
    From https://github.com/talkincode/ToughRADIUS
     * branch            master     -> FETCH_HEAD
    ...
    ...
    rad_console: stopped
    radiusd: stopped
    radiusd: started
    rad_console: started
    upgrade ok


配置文件修改
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

配置文件在/var/toughradius/radiusd.json

你可以修改其中的内容，你甚至可以指定另外的mysql数据库。

你应该修改 console的radaddr参数，改成你主机的IP地址。

如果你修改了端口，必须同时改变容器映射端口，你可以删除容器再重新创建。


删除容器::

    $ docker rm toughradius

重新创建容器时，只要没有删除/var/toughradius下的mysql目录数据文件，是不会重新创建和覆盖数据文件和配置文件的。
















