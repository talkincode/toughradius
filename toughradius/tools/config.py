#!/usr/bin/env python
#coding:utf-8
import ConfigParser
from toughradius.tools.shell import shell as sh
from toughradius.tools.secret import gen_secret

def setup_config():
    sh.info("set config...")
    config_path = sh.read('set your config file path,[ /etc/radiusd.conf ]') or '/etc/radiusd.conf'
    config = ConfigParser.RawConfigParser()
    sh.info("set default option")
    config.set('DEFAULT', 'debug', (sh.read("set debug [0/1] [0]:") or '0') )
    config.set('DEFAULT', 'tz', (sh.read("time zone [ CST-8 ]:") or 'CST-8') )
    config.set('DEFAULT','secret',gen_secret(32))
    
    sh.info("set database option")
    config.add_section('database')
    config.set('database','dbtype', (sh.read("database type [mysql]:") or 'mysql' ))
    config.set('database','host',( sh.read("database host [127.0.0.1]:") or '127.0.0.1' ))
    config.set('database','port',(sh.read("database port [3306]:") or '3306'))
    config.set('database','db',(sh.read("database dbname [toughradius]:") or 'toughradius' ))
    config.set('database','user',(sh.read("database user [root]:") or 'root' ))
    config.set('database','passwd',(sh.read("database passwd []:") or '' ))
    config.set('database','maxusage',(sh.read("db pool maxusage [30]:") or '30'))
    config.set('database','charset','utf8' )
    
    sh.info("set radiusd option")
    config.add_section('radiusd')
    config.set('radiusd','authport',(sh.read("radiusd authport [1812]:") or '1812'))
    config.set('radiusd','acctport',(sh.read("radiusd acctport [1813]:") or '1813'))
    config.set('radiusd','adminport',(sh.read("radiusd adminport [1815]:") or '1815'))
    config.set('radiusd','cache_timeout',(sh.read("radiusd cache_timeout (second) [600]:") or '600'))
    config.set('radiusd', 'logfile', (sh.read("log file [ logs/radiusd.log ]:") or 'logs/radiusd.log') )
    
    sh.info("set mysql backup ftpserver option")
    config.add_section('backup')
    config.set('backup','ftphost',(sh.read("backup ftphost [127.0.0.1]:") or '127.0.0.1' ))
    config.set('backup','ftpport',(sh.read("backup ftpport [21]:") or '21'))
    config.set('backup','ftpuser',(sh.read("backup ftpuser [ftpuser]:") or 'ftpuser' ))
    config.set('backup','ftppwd',(sh.read("backup ftppwd [ftppwd]:") or 'ftppwd' ))
    
    sh.info("set admin option")
    config.add_section('admin')
    config.set('admin','port',(sh.read("admin http port [1816]:") or '1816'))
    config.set('admin', 'logfile', (sh.read("log file [ logs/admin.log ]:") or 'logs/admin.log') )
    
    sh.info("set customer option")
    config.add_section('customer')
    config.set('customer','port',(sh.read("customer http port [1817]:") or '1817'))
    config.set('customer', 'logfile', (sh.read("log file [ logs/customer.log ]:") or 'logs/customer.log') )
    
    with open(config_path,'wb') as configfile:
        config.write(configfile)
        sh.succ("config save to %s"%config_path)
        
def echo_my_cnf():
    return '''[client]
#password=your_password
port=3306
socket=/var/toughradius/mysql/mysql.sock

[mysqld]
back_log=60
datadir=/var/toughradius/mysql
socket=/var/toughradius/mysql/mysql.sock
default-storage-engine=InnoDB
symbolic-links=0

wait_timeout=31536000
interactive_timeout=31536000

log-bin=mysql-bin
max_allowed_packet=32M

# explicit_defaults_for_timestamp

# Recommended in standard MySQL setup
sql_mode=NO_ENGINE_SUBSTITUTION,STRICT_TRANS_TABLES

innodb_data_home_dir=/var/toughradius/mysql
innodb_data_file_path=ibdata1:256M;ibdata2:256M:autoextend
innodb_buffer_pool_size=128M 
innodb_log_group_home_dir=/var/toughradius/mysql
innodb_additional_mem_pool_size=16M
innodb_log_file_size=32M
innodb_log_buffer_size=8M
innodb_flush_log_at_trx_commit=1
innodb_lock_wait_timeout=50
innodb_thread_concurrency=4

[mysqld_safe]
log-error=/var/toughradius/log/mysqld.log
pid-file=/var/toughradius/mysql/mysqld.pid    
'''

def echo_radiusd_cnf():
    return '''[DEFAULT]
debug = 0
tz = CST-8
secret = %s

[database]
dbtype = mysql
host = 127.0.0.1
port = 3306
db = toughradius
maxusage = 10
charset = utf8
user = root
passwd = 

[radiusd]
acctport = 1813
adminport = 1815
authport = 1812
cache_timeout = 600
logfile = /var/toughradius/log/radiusd.log

[admin]
port = 1816
logfile = /var/toughradius/log/admin.log

[customer]
port = 1817
logfile = /var/toughradius/log/customer.log

[backup]
ftpserver = 127.0.0.1
ftpport = 21
ftpuser = user
ftppwd = pwd 
'''%gen_secret(32)

def echo_supervisord_cnf():
    return '''[unix_http_server]
file=/tmp/supervisor.sock   ; (the path to the socket file)

[inet_http_server]         ; inet (TCP) server disabled by default
port=127.0.0.1:9001        ; (ip_address:port specifier, *:port for all iface)

[supervisord]
logfile=/var/toughradius/log/supervisord.log ; (main log file;default $CWD/supervisord.log)
logfile_maxbytes=64MB        ; (max main logfile bytes b4 rotation;default 50MB)
logfile_backups=4           ; (num of main logfile rotation backups;default 10)
loglevel=info                ; (log level;default info; others: debug,warn,trace)
pidfile=/tmp/supervisord.pid ; (supervisord pidfile;default supervisord.pid)
nodaemon=false               ; (start in foreground if true;default false)
minfds=1024                  ; (min. avail startup file descriptors;default 1024)
minprocs=200                 ; (min. avail process descriptors;default 200)

[rpcinterface:supervisor]
supervisor.rpcinterface_factory = supervisor.rpcinterface:make_main_rpcinterface

[supervisorctl]
;serverurl=unix:///tmp/supervisor.sock ; use a unix:// URL  for a unix socket
serverurl=http://127.0.0.1:9001 ; use an http:// url to specify an inet socket

[program:radiusd]
command=toughctl -radiusd 
process_name=%(program_name)s
numprocs=1
directory=/var/toughradius
autostart=true
autorestart=true
user=root
redirect_stderr=true
stdout_logfile=/var/toughradius/log/radiusd.log

[program:rad_console]
command=toughctl -admin 
process_name=%(program_name)s
numprocs=1
directory=/var/toughradius
autostart=true
autorestart=true
user=root
redirect_stderr=true
stdout_logfile=/var/toughradius/log/admin.log

[program:rad_customer]
command=toughctl -customer 
process_name=%(program_name)s
numprocs=1
directory=/var/toughradius
autostart=true
autorestart=true
user=root
redirect_stderr=true
stdout_logfile=/var/toughradius/log/customer.log
'''

def echo_centos7_service():
    return '''[Unit]  
Description=toughrad  
After=network.target remote-fs.target nss-lookup.target  
   
[Service]  
Type=forking  
PIDFile=/var/toughradius/toughrad.pid
ExecStart=toughrad start 
ExecReload=toughrad restart 
ExecStop=toughrad stop 
PrivateTmp=true  

[Install]  
WantedBy=multi-user.target
'''