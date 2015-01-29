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
}

mysql()
{
    echo "install mysql"
    rpm -ivh http://dev.mysql.com/get/mysql-community-release-el7-5.noarch.rpm
    yum install -y mysql-community-server mysql-community-devel 
}


radius()
{
    echo "pull ToughRADIUS latest"
    git clone https://github.com/talkincode/ToughRADIUS.git /opt/toughradius
    pip install -r /opt/toughradius/requirements.txt

    echo "mkdir /var/toughradius"
    mkdir -p /var/toughradius
    mkdir -p /var/toughradius/mysql
    mkdir -p /var/toughradius/log

    yes | cp -f /opt/toughradius/docker/my.cnf /etc/my.cnf
    yes | cp -f /opt/toughradius/docker/radiusd.json /var/toughradius/radiusd.json
    yes | cp -f /opt/toughradius/docker/supervisord.conf /var/toughradius/supervisord.conf    
    ln -s /opt/toughradius/install/upgrade.sh /usr/bin/radius_upgrade
    chmod +x /usr/bin/radius_upgrade
}

setup()
{
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
    echo "to delete toughradius "
    read -s -n1 -p "Press any key to continue ... "
    echo "shutdown mysql.."
    mysql -uroot shutdown
    echo "shutdown supervisord"
    supervisorctl shutdown
    echo "/opt/toughradius"
    rm -fr /opt/toughradius
    echo "/var/toughradius"
    rm -fr /var/toughradius
    echo "/usr/bin/radius_upgrade"
    rm -f /usr/bin/radius_upgrade
    echo 'unsetup done!'
}


case "$1" in

  depend)
    depend
  ;;

  mysql)
    mysql
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
  
  all)
    depend
    mysql
    radius
    setup
  ;;

  *)
    echo "Usage: $0 {depend|mysql|radius|setup|unsetup|all}"
  ;;

esac
