#!/bin/sh

start()
{
   echo "startup..."
   nohup radiusd >> /dev/null &
}

stop()
{
   echo "stop..."
   ps aux | grep "radiusd" | awk '{print $2}' | xargs kill -9

}

upgrade()
{
   pip install -U toughradius
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

    All other options are passed to the radiusd program.
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
    ps aux | grep "radiusd"
  ;;

  upgrade)
    upgrade
  ;;

  *)
   usage
  ;;
esac