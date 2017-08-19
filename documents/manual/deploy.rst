Deployment
=================

Supervisor
-------------------------

Supervisor deployment is recommended

To install supervisor

::

    pip install supervisor

The supervisor configuration file

::

    [unix_http_server]
    file=/tmp/toughradius.sock

    [inet_http_server]
    port=127.0.0.1:19001
    username=ctlman
    password=ctlroot

    [supervisord]
    nodaemon=false
    logfile=/etc/toughradius/toughradius.log
    logfile_maxbytes=1MB
    logfile_backups=8
    loglevel=debug
    pidfile=/etc/toughradius/toughradius.pid
    ;user=root

    [rpcinterface:supervisor]
    supervisor.rpcinterface_factory = supervisor.rpcinterface:make_main_rpcinterface

    [supervisorctl]
    serverurl=unix:///tmp/toughradius.sock

    [program:api]
    command=gtrad apiserv -c /etc/toughradius/radiusd.json
    dictionary=/etc/toughradius
    startretries = 5
    autorestart = true
    redirect_stderr=true
    stdout_logfile=/etc/toughradius/gtrad-api.log


    [program:auth]
    command=gtrad auth -c /etc/toughradius/radiusd.json
    dictionary=/etc/toughradius
    startretries = 5
    autorestart = true
    redirect_stderr=true
    stdout_logfile=/etc/toughradius/gtrad-auth.log

    [program:acct]
    command=gtrad acct -c /etc/toughradius/radiusd.json
    dictionary=/etc/toughradius
    startretries = 5
    autorestart = true
    redirect_stderr=true
    stdout_logfile=/etc/toughradius/gtrad-acct.log

Startup
~~~~~~~~~~~~~

::

    supervisord -c /etc/toughradius/radiusd.conf


Systemd
-------------

Systemd configuration file

::

    [Unit]
    Description=toughradius
    After=network.target

    [Service]
    Type=forking
    ExecStart=supervisord -c /etc/toughradius/radiusd.conf
    ExecReload=supervisorctl reload -c /etc/toughradius/radiusd.conf
    ExecStop=supervisorctl shutdown -c /etc/toughradius/radiusd.conf
    PrivateTmp=true

    [Install]
    WantedBy=multi-user.target

install service

::

    ln -s /etc/toughradius/toughradius.service /usr/lib/systemd/system/toughradius.service
    systemctl daemon-reload
