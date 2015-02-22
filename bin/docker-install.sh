#!/bin/sh
# =============================================================================
# jamiesun <jamiesun.net@gmail.com>
#
# CentOS-7, MySQL5.6, Python2.7, ToughRADIUS
# 
# =============================================================================

appdir=/usr/local/toughradius
rundir=/var/toughradius

depend()
{
    echo "install lib"
    yum update -y
    yum install -y wget git gcc tcpdump crontabs python-devel python-setuptools 
    echo "install mysql"
    yum install -y mariadb mariadb-server mariadb-devel MySQL-python
    echo "install python package"
    easy_install pip 
    pip install supervisor
    pip install DBUtils==1.1
    pip install Mako==0.9.0
    pip install Beaker==1.6.4
    pip install MarkupSafe==0.18
    pip install PyYAML==3.10
    pip install SQLAlchemy==0.9.8
    pip install Twisted==14.0.2
    pip install autobahn==0.9.3-3
    pip install bottle==0.12.7
    pip install six==1.8.0
    pip install tablib==0.10.0
    pip install zope.interface==4.1.1
    pip install pycrypto==2.6.1
    pip install sh==1.11
    pip install nose
    echo "install depend done!"
}


radius()
{
    echo "fetch ToughRADIUS latest"
    git clone -b stable https://github.com/talkincode/ToughRADIUS.git ${appdir}
    pip install -e ${appdir}
    echo "fetch ToughRADIUS done!"
}

setup()
{
    echo "mkdir /var/toughradius"
    mkdir -p ${rundir}
    mkdir -p ${rundir}/mysql
    mkdir -p ${rundir}/log
    
    toughctl --echo_my_cnf > ${rundir}/mysql/my.cnf
    toughctl --echo_radiusd_cnf > ${rundir}/radiusd.conf
    toughctl --echo_supervisord_cnf ${rundir}/supervisord.conf    

    chown -R mysql:mysql ${rundir}/mysql
    
    echo "starting install mysql database;"

    sleep 1s

    mysql_install_db --defaults-file=${rundir}/mysql/my.cnf --user=mysql --datadir=${rundir}/mysql

    /usr/bin/mysqld_safe --defaults-file=${rundir}/mysql/my.cnf --user=mysql &

    sleep 5s

    echo "GRANT ALL ON *.* TO admin@'%' IDENTIFIED BY 'radius' WITH GRANT OPTION; FLUSH PRIVILEGES" | mysql \
        --defaults-file=${rundir}/mysql/my.cnf

    echo "setup toughradius database.."

    toughctl -initdb -c ${rundir}/radiusd.conf 
    
    echo "add crontab task"
    
    echo '30 1 * * * $(which toughctl) --backup -c ${rundir}/radiusd.conf > /dev/null' > /tmp/backup.cron
    
    crontab /tmp/backup.cron

    echo "starting supervisord..."
    
    echo "ok" > ${rundir}/install.log

    supervisord -n -c ${rundir}/supervisord.conf
}

unsetup()
{
    echo "shutdown mysql.."
    mysqladmin --defaults-file=${rundir}/mysql/my.cnf -uroot shutdown
    echo "shutdown supervisord"
    supervisorctl shutdown
    echo ${rundir}
    rm -fr ${rundir}
    echo 'unsetup done!'
}

upgrade()
{
    echo 'starting upgrade...' 
    cd ${appdir} && git pull 
    supervisorctl restart all
    supervisorctl status
    echo 'upgrade done'
}


case "$1" in

  depend)
    depend
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
