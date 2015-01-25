#/bin/bash

echo 'starting upgrade...' 

cd /opt/toughradius \
    && git push origin master \
    && supervisorctl restart all

echo 'upgrade ok'