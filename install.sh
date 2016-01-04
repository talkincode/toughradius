#!/bin/sh
# toughradius v2.0 install script
# author: jamiesun.net@gmail.com

rundir=/home/toughrun

usage ()
{
    cat <<EOF
Usage: $0 [OPTIONS]
  docker             install docker, docker-compose
  standalone         install toughradius with already exists mysql
  with_mysql         install toughradius with new docker mysql instance
  remove             uninstall toughradius and database
All other options are passed to the toughrad program.
EOF
        exit 1
}

with_mysql()
{
    mkdir -p ${rundir}/toughradius

    read -p "mysql user [raduser]: " mysql_user
    mysql_user=${mysql_user:-raduser}

    read -p "mysql user password [radpwd]: " mysql_password
    mysql_password=${mysql_password:-radpwd}

    read -p "mysql database [radiusd]: " mysql_database
    mysql_database=${mysql_database:-radiusd}

    read -p "mysql root password [radroot]: " mysql_root_password
    mysql_root_password=${mysql_root_password:-radroot}

    read -p "toughradius web port [1816]: " web_port
    web_port=${web_port:-1816}

    read -p "toughradius auth port [1812]: " auth_port
    auth_port=${auth_port:-1812}

    read -p "toughradius acct port [1813]: " acct_port
    acct_port=${acct_port:-1813}

    cat <<EOF
    toughradius install config:

    mysql_user: ${mysql_user}
    mysql_password: ${mysql_password}
    mysql_database: ${mysql_database}
    mysql_root_password: ${mysql_root_password}
    web_port: ${web_port}
    auth_port: ${auth_port}
    acct_port: ${acct_port}
EOF

    cat > ${rundir}/toughradius/docker-compose.yml <<EOF 
raddb:
    image: "index.alauda.cn/toughstruct/mysql:512M"
    privileged: true
    expose:
        - "3306"
    environment:
        - MYSQL_USER=$mysql_user
        - MYSQL_PASSWORD=$mysql_password
        - MYSQL_DATABASE=$mysql_database
        - MYSQL_ROOT_PASSWORD=$mysql_root_password
    restart: always
    volumes:
        - ${rundir}/trmysql:/var/lib/mysql

radius:
    image: "index.alauda.cn/toughstruct/toughradius:v2"
    command: pypy /opt/toughradius/toughctl --standalone
    ports:
        - "${web_port}:${web_port}"
        - "${auth_port}:${auth_port}/udp"
        - "${acct_port}:${acct_port}/udp"
    links:
        - raddb:raddb
    environment:
        - DB_TYPE=mysql
        - DB_URL=mysql://$mysql_user:$mysql_password@raddb:3306/$mysql_database?charset=utf8
    restart: always
    volumes:
        - ${rundir}/toughradius:/var/toughradius
EOF 

    docker-compose up -d

    docker-compose ps

    exit 0
}

standalone()
{
    mkdir -p ${rundir}/toughradius

    read -p "mysql host (must): " mysql_host
    if [ -z $mysql_host ]; then
        echo "mysql host is empty"
        exit 1
    fi

    read -p "mysql user [raduser]: " mysql_user
    mysql_user=${mysql_user:-raduser}

    read -p "mysql user password [radpwd]: " mysql_password
    mysql_password=${mysql_password:-radpwd}

    read -p "mysql database [radiusd]: " mysql_database
    mysql_database=${mysql_database:-radiusd}

    read -p "mysql root password [radroot]: " mysql_root_password
    mysql_root_password=${mysql_root_password:-radroot}

    read -p "toughradius web port [1816]: " web_port
    web_port=${web_port:-1816}

    read -p "toughradius auth port [1812]: " auth_port
    auth_port=${auth_port:-1812}

    read -p "toughradius acct port [1813]: " acct_port
    acct_port=${acct_port:-1813}

    cat <<EOF
ToughRADIUS install config:

    mysql_host: ${mysql_host}
    mysql_user: ${mysql_user}
    mysql_password: ${mysql_password}
    mysql_database: ${mysql_database}
    mysql_root_password: ${mysql_root_password}
    web_port: ${web_port}
    auth_port: ${auth_port}
    acct_port: ${acct_port}
EOF

    cat > ${rundir}/toughradius/docker-compose.yml  <<EOF
radius:
    image: "index.alauda.cn/toughstruct/toughradius:v2"
    command: pypy /opt/toughradius/toughctl --standalone
    ports:
        - "${web_port}:${web_port}"
        - "${auth_port}:${auth_port}/udp"
        - "${acct_port}:${acct_port}/udp"
    environment:
        - DB_TYPE=mysql
        - DB_URL=mysql://$mysql_user:$mysql_password@raddb:3306/$mysql_database?charset=utf8
    restart: always
    volumes:
        - ${rundir}/toughradius:/var/toughradius
EOF

    cd ${rundir}/toughradius

    docker-compose up -d

    docker-compose ps

    exit 0
}


docker()
{
    curl -sSL https://get.daocloud.io/docker | sh

    curl -L https://get.daocloud.io/docker/compose/releases/download/1.5.2/docker-compose-`uname -s`-`uname -m` > /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose

    ln -s /usr/local/bin/docker-compose /usr/local/bin/docp
}

remove()
{
    cd ${rundir}/toughradius
    read -p "remove database [y/n](n): " rmdata
    if [ rmdata == 'y' ]; then
        docker-compose stop raddb
        docker-compose rm raddb
    fi 
    
    docker-compose stop radius
    docker-compose rm radius
}



case "$1" in

  docker)
    docker
  ;;

  standalone)
    standalone
  ;;

  with_mysql)
    with_mysql
  ;;

  remove)
    remove
  ;;

  *)
   usage
  ;;

esac