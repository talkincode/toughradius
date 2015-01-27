.. ToughRADIUS documentation master file, created by
   sphinx-quickstart on Thu Jan  8 16:28:26 2015.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.


ToughRADIUS手册
=======================================

简介
--------

ToughRADIUS是一个开源，免费，易用的Radius服务软件。

ToughRADIUS支持标准RADIUS协议，提供完整的AAA实现。支持灵活的策略管理，支持各种主流接入设备并轻松扩展，完美对接RouterOS，丰富的计费策略支持。

ToughRADIUS支持MySQL存储用户数据，并支持数据缓存，极大的提高了性能。

ToughRADIUS支持Windows，Linux跨平台部署，部署使用简单。

提供了RADIUS核心服务引擎与Web管理控制台,用户自助服务三个子系统，核心服务引擎提供高性能的认证计费服务，Web管理控制台提供了界面友好，功能完善的管理功能。用户自助服务系统提供了一个面向终端用户的网上服务渠道。

ToughRADIUS主站点:http://www.toughradius.net

ToughRADIUS交流社区:http://forum.toughradius.net

ToughRADIUS QQ交流群：247860313


.. _install-docs:

安装手册
--------

.. toctree::
    :maxdepth: 3

    windows_install
    linux_install
    docker_install


.. _management-docs:

系统管理
--------

.. toctree::
    :maxdepth: 3

    management/param
    management/node
    management/bas
    management/product


.. _business-docs:

营业管理
--------

.. toctree::
    :maxdepth: 3

    business/service
    business/accept
    
.. _operate-docs:

运维管理
--------

.. toctree::
    :maxdepth: 3

    operate/user_trace
    

.. _develop-docs:

开发手册
--------

.. toctree::
    :maxdepth: 3

    develop/database
    develop/dbdev
    

.. _bas-docs:

设备对接手册
------------

.. toctree::
    :maxdepth: 3

    case/routeros
    case/routeros_pppoe
    case/routeros_attrs


