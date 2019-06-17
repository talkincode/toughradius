## 系统环境依赖

- 操作系统：Linux(推荐CentOS7)
- java 版本: 1.8
- 数据库服务器：MySQL/MariaDB

## 上传安装包到服务器

通过sftp或ftp上传安装包到服务器目录并解压， 通过终端 cd 进入解压目录， 比如

> /opt/toughradius-v6.1.1.3

## 数据库初始化

> 数据库的安装配置请自行完成,首先确保你的数据库服务器已经运行

执行安装目录下的 installer.sh 脚本进行初始化数据库

> sh installer.sh initdb

## 安装服务程序

> sh installer.sh install

## 修改配置

注意修改 /opt/application-prod.properties 配置文件中的数据库部分

如果希望使用自定义的模板，请取消该行注释

> `#org.toughradius.portal.templateDir=file:/opt/portal/`

/opt/portal/ 是自定义模板目录， 可以参照安装包里的模板进行修改

## 运行服务

> systemctl start toughradius