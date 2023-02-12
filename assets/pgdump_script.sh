#!/bin/bash

DB_HOST="{dbhost}"
DB_PORT="{dbport}"
DB_USER="{dbuser}"
DB_PWD="{dbpwd}"
DB_NAME="{dbname}"
backupdate=`date  +"%Y%m%d-%H%M%S"`
MAX_DAYS=14
BACKUPS_DIR="/var/toughradius/backup/database"
PSQ="/usr/bin/psql"
PGDUMP="/usr/bin/pg_dump"
VACUUM="/usr/bin/vacuumdb"


echo "${DB_HOST}:${DB_PORT}:${DB_NAME}:${DB_USER}:${DB_PWD}" > ~/.pgpass
chmod 600 ~/.pgpass

####log_correct函数打印正确的输出到日志文件
function log_correct () {
  DATE=`date +"%Y%m%d-%H%M%S"`  ####显示打印日志的时间
  USER=$(whoami) ####那个用户在操作
  echo "${DATE} ${USER} execute $0 [INFO] $@" >> log_info.log ######（$0脚本本身，$@将参数作为整体传输调用）
}
#log_error打印shell脚本中错误的输出到日志文件
function log_error ()
{
  DATE=`date +"%Y%m%d-%H%M%S"`
  USER=$(whoami)
  echo "${DATE} ${USER} execute $0 [INFO] $@" >> log_error.log ######（$0脚本本身，$@将参数作为整体传输调用）
}
####fn_log函数 通过if判断执行命令的操作是否正确，并打印出相应的操作输出
function fn_log ()
{
  if [ $? -eq 0 ]
  then
    log_correct "$@ sucessed!"
    echo -e "\033[32m $@ sucessed. \033[0m"
  else
    log_error "$@ failed!"
    echo -e "\033[41;37m $@ failed. \033[0m"
    exit
  fi
}

echo -n "Vacuuming database ${DB_NAME} ... "
$VACUUM -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -z ${DB_NAME}
log_correct "${DB_NAME} update optimizer statistics Done."
BACKUP_NAME="$BACKUPS_DIR/${DB_NAME}-$backupdate.sql.gz"
log_correct "Creating backup of database ${DB_NAME}"
$PGDUMP -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} \
 --disable-triggers \
 --data-only \
 --exclude-table="ts_*" \
 --exclude-table="ts_radius_accounting" \
 --exclude-table-data='_timescaledb_internal.*' \
 --exclude-table-data='_timescaledb_catalog.*' \
 --exclude-table-data='_timescaledb_cache.*' \
 --exclude-table-data='_timescaledb_config.*' \
 --no-password \
 -w ${DB_NAME}  -Ft | gzip > ${BACKUP_NAME}


log_correct "${DB_NAME} backup Done."
echo -n "Removing backups of database ${DB_NAME} older than $MAX_DAYS days"
find $BACKUPS_DIR  -mtime +$MAX_DAYS  -type f -name "${DB_NAME}*" -exec rm -f {} \;
log_correct "Removing backups of database ${DB_NAME} older than $MAX_DAYS days."