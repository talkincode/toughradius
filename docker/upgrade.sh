#/bin/bash

echo 'starting upgrade...' 

cd /opt/toughradius \
    && git pull origin master

yes | cp -f /opt/toughradius/docker/startup.sh /opt/startup.sh

chmod +x /opt/startup.sh
supervisorctl restart all

echo 'upgrade ok'