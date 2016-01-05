# TOUGHRADIUS

[![](https://badge.imagelayers.io/talkincode/toughradius:v2.svg)](https://imagelayers.io/?images=talkincode/toughradius:v2 'Get your own badge on imagelayers.io')

## TOUGHRADIUS 简介

TOUGHRADIUS是一个开源的Radius服务软件，采用于AGPL许可协议发布。

TOUGHRADIUS支持标准RADIUS协议，提供完整的AAA实现。支持灵活的策略管理，支持各种主流接入设备并轻松扩展，具备丰富的计费策略支持。

TOUGHRADIUS支持使用Oracle, MySQL, PostgreSQL, MSSQL等主流数据库存储用户数据，并支持数据缓存，极大的提高了性能。

TOUGHRADIUS支持Windows，Linux，BSD跨平台部署，部署使用简单。

TOUGHRADIUS提供了RADIUS核心服务引擎与Web管理控制台,用户自助服务三个子系统，核心服务引擎提供高性能的认证计费服务，Web管理控制台提供了界面友好，功能完善的管理功能。用户自助服务系统提供了一个面向终端用户的网上服务渠道。

TOUGHRADIUS网站：http://www.toughradius.net

## Linux 快读部署

TOUGHRADIUS 提供了一个Linux 工具脚本，可以实现TOUGHRADIUS的部署与管理

    wget  https://github.com/talkincode/ToughRADIUS/raw/master/trshell  -O /usr/local/bin/trshell

    chmod +x /usr/local/bin/trshell


- 安装docker环境

    trshell docker_setup


- 一键部署 TOUGHRADIUS, 同时部署一个MySQL实例

    trshell with_mysql t1     # t1表示实例名，可自定义


- 一键部署TOUGHRADIUS, 连接已有的MySQL数据库

    trshell standalone t1     ＃ t1表示实例名，可自定义


访问 http://server:1816  进入管理系统，默认用户名密码是 admin/root

## TOUGHRADIUS 文档

http://docs.toughradius.net

## TOUGHRADIUS 商业授权

[TOUGHRADIUS 商业授权](#) (https://github.com/talkincode/ToughRADIUS/blob/master/Commerical-license.rst)