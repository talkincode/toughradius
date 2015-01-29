#/bin/bash

echo 'starting upgrade...' 

cd /opt/toughradius && git pull origin master
supervisorctl restart all
sleep 1s
supervisorctl status

echo 'upgrade done'