#!/usr/bin/env python
#coding:utf-8
from __future__ import unicode_literals
import datetime
import shutil
import os
import sys

sysctlstr = '''
net.ipv4.ip_forward=1
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_tw_recycle = 1
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_keepalive_time = 1200
net.ipv4.ip_local_port_range = 10000 65000
net.ipv4.tcp_max_syn_backlog = 8192
net.ipv4.tcp_max_tw_buckets = 5000
net.core.netdev_max_backlog = 32768
net.core.somaxconn = 32768
net.core.wmem_default = 33554432
net.core.rmem_default = 33554432
net.core.rmem_max = 134217728
net.core.wmem_max = 134217728
net.ipv4.tcp_synack_retries = 2
net.ipv4.tcp_syn_retries = 2
net.ipv4.tcp_wmem = 8192 436600 873200
net.ipv4.tcp_rmem  = 32768 436600 873200
net.ipv4.tcp_mem = 94500000 91500000 92700000
net.ipv4.tcp_max_orphans = 3276800
vm.overcommit_memory = 1
'''

limitstr = '''
# start upcore_limit_config

*  soft nproc 65535
*  hard nproc 65535
*  soft nofile 65535
*  hard nofile 65535

mysql  soft nproc 65535
mysql  hard nproc 65535
mysql  soft nofile 65535
mysql  hard nofile 65535

# end upcore_limit_config
'''

loginstr = '''
# start upcore_limit_config

session required /lib64/security/pam_limits.so
session required pam_limits.so

# end upcore_limit_config
'''

profilestr = 'ulimit -n 65535'


def upcore():
    # backup sysconfig
    backup_time = datetime.datetime.now().strftime("%Y%m%d%H%M%S")
    shutil.copy('/etc/sysctl.conf','/etc/sysctl.conf.bak.%s'%backup_time)
    shutil.copy('/etc/security/limits.conf','/etc/security/limits.conf.bak.%s'%backup_time)
    shutil.copy('/etc/pam.d/login','/etc/pam.d/login.bak.%s'%backup_time)


    with open('/etc/sysctl.conf','wb') as sysfs:
        sysfs.write(sysctlstr)

    os.system("sysctl -p")

    is_limit_up = False
    with open("/etc/security/limits.conf",'rb') as limitfs:
        for line in limitfs:
            if 'start upcore_limit_config' in line:
                is_limit_up = True
                break

    with open( "/etc/security/limits.conf", 'wab' ) as limitfsa:
        limitfsa.write(limitstr)

    os.system("cat /etc/security/limits.conf")

    is_login_up = False
    with open("/etc/pam.d/login",'rb') as loginfs:
        for line in loginfs:
            if 'start upcore_limit_config' in line:
                is_login_up = True
                break

    with open( "/etc/pam.d/login", 'wab' ) as loginfsa:
        loginfsa.write(loginstr)

    os.system( "cat /etc/pam.d/login" )


def up_mariadb():
    if not os.path.exists("/etc/systemd/system/mariadb.service.d"):
        os.makedirs("/etc/systemd/system/mariadb.service.d")

    with open('/etc/systemd/system/mariadb.service.d/limits.conf','wb') as mfs:
        mfs.write("[Service]\nLimitNOFILE=65535\n")

    os.system("systemctl daemon-reload")
    os.system("cat /etc/systemd/system/mariadb.service.d/limits.conf")


def install_toughradius():
    cfgparam = dict(
        server_port = raw_input('请输入web服务端口 (默认 1816)'.encode("utf-8")) or 1816,
        auth_port = raw_input('请输入 RADIUS 认证端口 (默认 1812)'.encode("utf-8")) or 1812,
        acct_port = raw_input('请输入 RADIUS 记账端口 (默认 1813)'.encode("utf-8")) or 1813,
        dbuser = raw_input('请输入数据库用户名 (默认 raduser)'.encode("utf-8")) or "raduser",
        dbpwd = raw_input('请输入数据库密码 (默认 radpwd)'.encode("utf-8")) or "radpwd",
        usessl = raw_input('是否启用 https ( true|false 默认 false)'.encode("utf-8")) or "false",
    )

    if raw_input('是否需要创建数据库 (y/n 默认 n)?'.encode("utf-8")) == 'y':
        os.system("curl -L http://115.159.56.13:8008/toughradius/latest/database.sql -O  /tmp/database.sql")
        os.system('mysql -uroot -p -e "create database toughradius DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" -v')
        os.system("""mysql -uroot -p -e "GRANT ALL ON toughradius.* TO raduser@'127.0.0.1' IDENTIFIED BY 'radpwd' WITH GRANT OPTION;FLUSH PRIVILEGES;" -v""")
        os.system('mysql -uroot -p  < /tmp/database.sql')

    config_fstr='''server.port = {server_port}
server.security.require-ssl={usessl}
logging.config=classpath:logback-prod.xml

# database config
spring.datasource.url=jdbc:mysql://127.0.0.1:3306/toughradius?serverTimezone=Asia/Shanghai&useUnicode=true&characterEncoding=utf-8&allowMultiQueries=true
spring.datasource.username={dbuser}
spring.datasource.password={dbpwd}
spring.datasource.max-active=120
spring.datasource.driver-class-name=com.mysql.cj.jdbc.Driver
'''.format(**cfgparam)

    with open('/opt/application-prod.properties','wb') as mfs:
        mfs.write(config_fstr)

    usememary = raw_input('请输入服务进程使用的最大内存 ( 默认 1024M )'.encode("utf-8")) or "1024M"
    service_fstr = '''[Unit]
Description=toughradius
After=syslog.target

[Service]
WorkingDirectory=/opt
User=root
LimitNOFILE=65535
LimitNPROC=65535
Type=simple
ExecStart=/usr/bin/java -jar -Xms{usememary} -Xmx{usememary} /opt/toughradius-latest.jar  --spring.profiles.active=prod
SuccessExitStatus=143

[Install]
WantedBy=multi-user.target'''.format(usememary=usememary)
    with open('/usr/lib/systemd/system/toughradius.service','wb') as mfs:
        mfs.write(service_fstr)
    os.system("systemctl enable toughradius")
    os.system("curl -L http://115.159.56.13:8008/toughradius/latest/toughradius-latest.jar -o /opt/toughradius-latest.jar")

    isrun = raw_input('安装完成，是否立即启动 (y/n 默认 y)?'.encode("utf-8")) or 'y'
    if isrun == 'y':
        os.system("systemctl start toughradius && systemctl status toughradius")


if __name__ == "__main__":
    usage= """
====================================================================================
    
##### 即将开始安装 toughradius，请仔细阅读以下内容，
    
    toughradius v6 基于LGPL V3协议发布， 继续安装代表您同意该协议
    
    LICENSE: https://raw.githubusercontent.com/talkincode/ToughRADIUS/master/LICENSE

    toughradius v6 依赖java 与 mysql环境， java 和 mariadb 需要自行安装，centos7 下 请参考如下指令：

    1 安装 java
    yum install -y java

    2 测试 java 安装是否有效
    java -version

    3 安装 mariadb|mysql
    yum install -y mariadb-server
    
    4 启动数据库，如果需要修改优化 mysql 配置， 请修改 /etc/my.cnf
    systemctl enable mariadb
    systemctl start mariadb

    5 安装过程会自动创建数据库,,如需手工初始化数据库请执行如下指令， 如需修改数据库用户名密码，
    curl -L http://115.159.56.13:8008/toughradius/latest/database.sql -O  /tmp/database.sql
    mysql -uroot -p -e "create database toughradius DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" -v
    mysql -uroot -p -e "GRANT ALL ON toughradius.* TO raduser@'127.0.0.1' IDENTIFIED BY 'radpwd' WITH GRANT OPTION;FLUSH PRIVILEGES;" -v
    mysql -uroot -p  < /tmp/database.sql

    6 注意还要开启防火墙端口，实际端口根据配置确定
    firewall-cmd --zone=public --add-port=1816/tcp --permanent
    firewall-cmd --zone=public --add-port=1812/udp --permanent
    firewall-cmd --zone=public --add-port=1813/udp --permanent

    7 安装完成后，如需修改程序配置， 请修改 /opt/application-prod.properties
       
====================================================================================
"""
    print(usage)

    if not raw_input('是否继续安装? (y/n)?'.encode("utf-8")) == 'y':
        sys.exit(0)

    if not os.path.exists("/usr/bin/java"):
        print("请安装 java 8")
        sys.exit(0)

    if not os.path.exists("/usr/bin/mysql"):
        print("请安装数据库 mariadb 或 mysql")
        sys.exit(0)

    if raw_input('是否需要优化系统内核?(y/n 默认 n)?'.encode("utf-8")) == 'y':
        upcore()

    if raw_input('是否需要优化mysql连接数限制?(y/n 默认 n)?'.encode("utf-8")) == 'y':
        up_mariadb()

    if raw_input('是否需要增加 ulimit -n 65535 到 /etc/profile (y/n 默认 n)?'.encode("utf-8")) == 'y':
        os.system("echo 'ulimit -n 65535' >> /etc/profile")


    install_toughradius()


