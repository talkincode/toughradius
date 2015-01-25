#/bin/bash

echo 'starting upgrade...' 

cd /opt/toughradius \
    && git pull origin master \
    && supervisorctl restart all

echo 'upgrade ok'