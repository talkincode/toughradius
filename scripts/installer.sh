#!/bin/sh

install_package()
{
    \cp application-prod.properties /opt/application-prod.properties
    \cp toughradius-latest.jar /opt/toughradius-latest.jar
    \cp toughradius.service /usr/lib/systemd/system/toughradius.service
    \cp -r portal /opt/
    systemctl enable toughradius
    echo "install done, please exec systenctl start toughradius after initdb"
}

setup_mysql()
{
    echo "create database toughradius"
    mysql -uroot -p -e "create database toughradius DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
    echo "GRANT db user"
    mysql -uroot -p -e "GRANT ALL ON toughradius.* TO raduser@'127.0.0.1' IDENTIFIED BY 'radpwd' WITH GRANT OPTION;FLUSH PRIVILEGES;" -v
    echo "create tables"
    mysql -uroot -p  < database.sql
    echo "insert test data"
    mysql -uroot -p  < init.sql
}

case "$1" in

  initdb)
    setup_mysql
  ;;

  install)
    install_package
  ;;

esac