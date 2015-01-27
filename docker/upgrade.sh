#/bin/bash

echo 'starting upgrade...' 

cd /opt/toughradius \
    && git pull origin master

yes | cp -f /opt/toughradius/docker/startup.sh /opt/startup.sh
yes | cp -f /opt/toughradius/docker/upgrade.sh /opt/upgrade.sh

chmod +x /opt/startup.sh
chmod +x /opt/upgrade.sh
supervisorctl restart all

echo 'upgrade done'