#!/usr/bin/env bash


case "$1" in

  info100)
    tail -n100 /var/toughradius/logs/toughradius.`date +%Y%m%d`.log
  ;;

  info1000)
    tail -n1000 /var/toughradius/logs/toughradius.`date +%Y%m%d`.log
  ;;

  err100)
    tail -n100 /var/toughradius/logs/toughradius-error.`date +%Y%m%d`.log
  ;;

  err1000)
    tail -n1000 /var/toughradius/logs/toughradius.error.`date +%Y%m%d`.log
  ;;

esac