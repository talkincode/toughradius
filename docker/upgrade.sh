#/bin/bash

echo 'starting upgrade...' 

cd /opt/toughradius \
    && git pull origin master

yes | cp -f /opt/toughradius/docker/startup.sh /opt/startup.sh
yes | cp -f /opt/toughradius/docker/upgrade.sh /opt/upgrade.sh
yes | cp -f /opt/toughradius/docker/radiusd.json /var/toughradius/radiusd.json
yes | cp -f /opt/toughradius/docker/supervisord.conf /var/toughradius/supervisord.conf

chmod +x /opt/startup.sh
chmod +x /opt/upgrade.sh

supervisorctl reload
supervisorctl restart all
supervisorctl status

echo 'upgrade done'