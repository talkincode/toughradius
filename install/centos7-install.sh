#!/bin/sh
# =============================================================================
# jamiesun <jamiesun.net@gmail.com>
#
# CentOS-7, MySQL5.6, Python2.7, ToughRADIUS
# 
# =============================================================================

appdir=/opt/toughradius
rundir=/var/toughradius

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
    git clone https://github.com/talkincode/ToughRADIUS.git ${appdir}
    pip install -r ${appdir}/requirements.txt
    echo "fetch ToughRADIUS done!"
}

setup()
{
    echo "mkdir /var/toughradius"
    mkdir -p ${rundir}
    mkdir -p ${rundir}/mysql
    mkdir -p ${rundir}/log
    
    yes | cp -f ${appdir}/install/my.cnf ${rundir}/mysql/my.cnf
    yes | cp -f ${appdir}/install/radiusd.json ${rundir}/radiusd.json
    yes | cp -f ${appdir}/install/supervisord.conf ${rundir}/supervisord.conf    
    yes | cp -f ${appdir}/install/toughrad.service /usr/lib/systemd/system/toughrad.service
    chmod 754 /usr/lib/systemd/system/toughrad.service
    ln -s ${appdir}/toughrad /usr/bin/toughrad 
    chmod +x /usr/bin/toughrad
    
    chown -R mysql:mysql ${rundir}/mysql
    
    echo "starting install mysql database;"

    sleep 1s

    mysql_install_db --defaults-file=${rundir}/mysql/my.cnf --user=mysql --datadir=${rundir}/mysql

    /usr/bin/mysqld_safe --defaults-file=${rundir}/mysql/my.cnf --user=mysql &

    sleep 5s

    echo "GRANT ALL ON *.* TO admin@'%' IDENTIFIED BY 'radius' WITH GRANT OPTION; FLUSH PRIVILEGES" | mysql \
        --defaults-file=${rundir}/mysql/my.cnf

    echo "setup toughradius database.."

    python ${appdir}/createdb.py -c ${rundir}/radiusd.json -i=1

    echo "starting supervisord..."

    supervisord -c ${rundir}/supervisord.conf

    sleep 3s

    echo "ToughRADIUS service status"

    supervisorctl status
    
    echo "setup done!"
}

unsetup()
{
    echo "shutdown mysql.."
    mysqladmin --defaults-file=${rundir}/mysql/my.cnf -uroot shutdown
    echo "shutdown supervisord"
    supervisorctl shutdown
    echo ${rundir}
    rm -fr ${rundir}
    rm -f /usr/bin/toughrad
    rm -f /usr/lib/systemd/system/toughrad.service
    echo 'unsetup done!'
}

upgrade()
{
    echo 'starting upgrade...' 
    cd ${appdir} && git pull origin master
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
