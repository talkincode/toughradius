.. ToughRADIUS documentation master file, created by
   sphinx-quickstart on Thu Jan  8 16:28:26 2015.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.


ToughRADIUS手册
=======================================

ToughRADIUS是一个开源，免费，易用的Radius服务软件。

ToughRADIUS支持标准RADIUS协议，提供完整的AAA实现。支持灵活的策略管理，支持各种主流接入设备并轻松扩展，完美对接RouterOS，丰富的计费策略支持。

ToughRADIUS支持MySQL存储用户数据，并支持数据缓存，极大的提高了性能。

ToughRADIUS支持Windows，Linux跨平台部署，部署使用简单。

提供了RADIUS核心服务引擎与Web管理控制台两个子系统，核心服务引擎提供高性能的认证计费服务，Web管理控制台提供了界面友好，功能完善的管理功能。

+ 关于AAA的概念

    AAA是Authentication（认证）、Authorization（授权）和Accounting（计费）的简称。它提供对用户进行认证、授权和计费三种安全功能。具体如下：

    - 1. 认证（Authentication）：认证用户是否可以获得访问权，确定哪些用户可以访问网络。
    - 2. 授权（Authorization）：授权用户可以使用哪些服务。
    - 3. 计费（Accounting）：记录用户使用网络资源的情况。

+ RADIUS协议

    RADIUS（Remote Authentication Dial In User Service）协议是在IETF的RFC2865和2866中定义的。RADIUS 是基于 UDP 的一种客户机/服务器协议。RADIUS客户机是网络访问服务器，它通常是一个路由器、交换机或无线访问点。RADIUS是AAA的一种实现协议。



.. _install-docs:

安装手册
--------

.. toctree::
    :maxdepth: 3

    windows_install


.. _management-docs:

管理手册
--------

.. toctree::
    :maxdepth: 3

    management/param
    management/node
    management/bas
    management/product




