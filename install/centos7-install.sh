#!/bin/sh
# =============================================================================
# jamiesun <jamiesun.net@gmail.com>
#
# CentOS-7, MySQL5.6, Python2.7, ToughRADIUS
# 
# =============================================================================


depend()
{
    echo "install lib"
    yum update -y
    yum install -y wget git gcc python-devel python-setuptools tcpdump

    echo "python package"
    easy_install pip 
    easy_install supervisor
    echo "install depend done!"
}

mysql5()
{
    echo "install mysql"
    rpm -ivh http://dev.mysql.com/get/mysql-community-release-el7-5.noarch.rpm
    yum install -y mysql-community-server mysql-community-devel 
    echo "install mysql done!"
}


radius()
{
    echo "fetch ToughRADIUS latest"
    git clone https://github.com/talkincode/ToughRADIUS.git /opt/toughradius
    pip install -r /opt/toughradius/requirements.txt
    echo "fetch ToughRADIUS done!"
}

setup()
{
    echo "mkdir /var/toughradius"
    mkdir -p /var/toughradius
    mkdir -p /var/toughradius/mysql
    mkdir -p /var/toughradius/log

    cp /etc/my.cnf /etc/my.cnf.$((`date +%s`))
    yes | cp -f /opt/toughradius/docker/my.cnf /etc/my.cnf
    yes | cp -f /opt/toughradius/docker/radiusd.json /var/toughradius/radiusd.json
    yes | cp -f /opt/toughradius/docker/supervisord.conf /var/toughradius/supervisord.conf    
    
    echo "starting install mysql database;"

    sleep 1s

    mysql_install_db 

    chown -R mysql:mysql /var/toughradius/mysql

    /usr/bin/mysqld_safe &

    sleep 5s

    echo "GRANT ALL ON *.* TO admin@'%' IDENTIFIED BY 'radius' WITH GRANT OPTION; FLUSH PRIVILEGES" | mysql

    echo "setup toughradius database.."

    python /opt/toughradius/createdb.py -c /var/toughradius/radiusd.json -i=1

    echo "starting supervisord..."

    supervisord -c /var/toughradius/supervisord.conf

    sleep 3s

    echo "ToughRADIUS service status"

    supervisorctl status
    
    echo "setup done!"
}

unsetup()
{
    echo "shutdown mysql.."
    mysqladmin -uroot shutdown
    echo "shutdown supervisord"
    supervisorctl shutdown
    echo "/var/toughradius"
    rm -fr /var/toughradius
    echo 'unsetup done!'
}

upgrade()
{
    echo 'starting upgrade...' 
    cd /opt/toughradius && git pull origin master
    supervisorctl restart all
    supervisorctl status
    echo 'upgrade done'
}


case "$1" in

  depend)
    depend
  ;;

  mysql5)
    mysql5
  ;;

  radius)
    radius
  ;;
  
  setup)
    setup
  ;;
  
  unsetup)
    unsetup
  ;;    
  
  upgrade)
    upgrade
  ;;    
  
  all)
    depend
    mysql5
    radius
    setup
  ;;

  *)
    echo "Usage: $0 {all|depend|mysql5|radius|setup|unsetup|upgrade}"
  ;;

esac
