#/bin/bash

if [ ! -f /var/toughradius/mysql/ibdata1 ]; then

    mkdir -p /var/toughradius/mysql
    mkdir -p /var/toughradius/log

    yes | cp -f /opt/toughradius/docker/radiusd.json /var/toughradius/radiusd.json
    yes | cp -f /opt/toughradius/docker/supervisord.conf /var/toughradius/supervisord.conf

    echo "starting install mysql database;"

    sleep 1s

    mysql_install_db

    chown -R mysql:mysql /var/toughradius/mysql

    /usr/bin/mysqld_safe &

    sleep 5s

    echo "GRANT ALL ON *.* TO admin@'%' IDENTIFIED BY 'radius' WITH GRANT OPTION; FLUSH PRIVILEGES" | mysql

    echo "create database test;" | mysql

    python /opt/toughradius/createdb.py -c /var/toughradius/radiusd.json -i=1

    mysqladmin -uroot shutdown
    sleep 1s

fi

echo "starting mysqd..."

/usr/bin/mysqld_safe &

sleep 7s

echo "starting supervisord..."

supervisord -n -c /var/toughradius/supervisord.conf



