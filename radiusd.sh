#!/bin/sh

pypy=/opt/toughcloud/venv/bin/pypy

start()
{
   echo "startup..."
   nohup $pypy /opt/toughcloud/toughcloud.py --workers=2 --web-port=1879 --cmd=runserver >> /dev/null &
   nohup $pypy /opt/toughcloud/toughcloud.py --workers=4 --web-port=18791 --cmd=runserver >> /dev/null &
   nohup $pypy /opt/toughcloud/toughcloud.py --cmd=sched >> /dev/null &
}

stop()
{
   echo "stop..."
   ps aux | grep "/opt/toughcloud/toughcloud.py" | awk '{print $2}' | xargs kill -9

}

upgrade()
{
   cd /opt/toughcloud
   git pull origin master
   stop
   start
}

usage ()
{
    cat <<EOF
    Usage: $0 [OPTIONS]

    start                startup application
    restart              restart application
    stop                 stop application
    upgrade              upgrade application

    All other options are passed to the toughcloud program.
EOF
        exit 1
}

case "$1" in

  start)
    start
  ;;

  stop)
    stop
  ;;

  restart)
    stop
    start
  ;;

  status)
    ps aux | grep "/opt/toughcloud/toughcloud.py"
  ;;

  upgrade)
    upgrade
  ;;

  *)
   usage
  ;;
esac