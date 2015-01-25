# ToughRADIUS

ToughRADIUS是一个开源，免费，易用的Radius服务软件。

ToughRADIUS支持标准RADIUS协议，提供完整的AAA实现。支持灵活的策略管理，支持各种主流接入设备并轻松扩展，完美对接RouterOS，支持包月，时长计费。

ToughRADIUS支持MySQL存储用户数据，支持数据缓存，极大的提高了性能。

ToughRADIUS支持Windows，Linux跨平台部署，部署使用简单。

提供了RADIUS核心服务引擎与Web管理控制台两个子系统，核心服务引擎提供高性能的认证计费服务，Web管理控制台提供了界面友好，功能完善的管理功能。

[ToughRADIUS网站：http:://www.toughradius.net] (http:://www.toughradius.net)

## 使用Docker镜像 

在centos7下部署::

    yum install docker

    service docker start

    mkdir /var/toughradius

    docker run -d -v /var/toughradius:/var/toughradius \
      -p 3306:3306 -p 1812:1812/udp -p 1813:1813/udp \
      -p 1815:1815 -p 1816:1816 \
      --name toughradius talkincode/centos7-toughradius 

以上指令会在centos7中安装docker工具，并自动下载toughradius镜像以守护进程模式运行。

运行 docker logs toughradius 即可看到运行日志输出。

打开浏览器访问 http://serverip:1816,可以进入web管理登陆界面了。


## 关于AAA的概念
    
AAA是Authentication（认证）、Authorization（授权）和Accounting（计费）的简称。它提供对用户进行认证、授权和计费三种安全功能。具体如下：
    
1. 认证（Authentication）：认证用户是否可以获得访问权，确定哪些用户可以访问网络。
2. 授权（Authorization）：授权用户可以使用哪些服务。
3. 计费（Accounting）：记录用户使用网络资源的情况。

## RADIUS协议
    
RADIUS（Remote Authentication Dial In User Service）协议是在IETF的RFC2865和2866中定义的。RADIUS 是基于 UDP 的一种客户机/服务器协议。RADIUS客户机是网络访问服务器，它通常是一个路由器、交换机或无线访问点。RADIUS是AAA的一种实现协议。