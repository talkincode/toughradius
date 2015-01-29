#!/bin/sh
# =============================================================================
# jamiesun <jamiesun.net@gmail.com>
#
# CentOS-7, MySQL5.6, Python2.7, ToughRADIUS
# 
# =============================================================================

echo "install lib"
yum update -y
yum install -y wget git gcc python-devel python-setuptools tcpdump

echo "install mysql"
rpm -ivh http://dev.mysql.com/get/mysql-community-release-el7-5.noarch.rpm
yum install -y mysql-community-server mysql-community-devel 

echo "python package"
easy_install pip 
easy_install supervisor

echo "pull ToughRADIUS latest"
git clone https://github.com/talkincode/ToughRADIUS.git /opt/toughradius
pip install -r /opt/toughradius/requirements.txt

if [ ! -d /var/toughradius ]; then
    echo "mkdir /var/toughradius"
    mkdir -p /var/toughradius
    mkdir -p /var/toughradius/mysql
    mkdir -p /var/toughradius/log

    yes | cp -f /opt/toughradius/docker/my.cnf /var/toughradius/my.cnf
    yes | cp -f /opt/toughradius/docker/radiusd.json /var/toughradius/radiusd.json
    yes | cp -f /opt/toughradius/docker/supervisord.conf /var/toughradius/supervisord.conf    
fi

if [ ! -f /var/toughradius/mysql/ibdata1 ]; then

    echo "starting install mysql database;"

    sleep 1s

    mysql_install_db --defaults-file=/var/toughradius/my.cnf

    chown -R mysql:mysql /var/toughradius/mysql

    /usr/bin/mysqld_safe &

    sleep 5s

    echo "GRANT ALL ON *.* TO admin@'%' IDENTIFIED BY 'radius' WITH GRANT OPTION; FLUSH PRIVILEGES" | mysql

    python /opt/toughradius/createdb.py -c /var/toughradius/radiusd.json -i=1

    mysqladmin -uroot shutdown
    sleep 1s
fi

echo "starting mysqld_safe..."

/usr/bin/mysqld_safe --defaults-file=/var/toughradius/my.cnf &

sleep 5s

echo "starting supervisord..."

supervisord -c /var/toughradius/supervisord.conf

sleep 3s

echo "ToughRADIUS service status"

supervisorctl status
