FROM index.alauda.cn/toughstruct/tough-pypy
MAINTAINER jamiesun <jamiesun.net@gmail.com>

RUN git clone -b stable https://github.com/talkincode/ToughRADIUS.git /opt/toughradius

RUN ln -s /opt/toughradius/toughctl /usr/bin/toughctl && chmod +x /usr/bin/toughctl

RUN ln -s /opt/toughradius/docker/radiusd.conf /etc/radiusd.conf
RUN ln -s /opt/toughradius/docker/supervisord.conf /etc/supervisord.conf
RUN ln -s /opt/toughradius/docker/toughrad /usr/bin/toughrad && chmod +x /usr/bin/toughrad
RUN ln -s /opt/toughradius/docker/privkey.pem /var/toughradius/privkey.pem
RUN ln -s /opt/toughradius/docker/cacert.pem /var/toughradius/cacert.pem

RUN pypy /opt/toughradius/toughctl --initdb

EXPOSE 1815 1816 1817 1819
EXPOSE 1812/udp
EXPOSE 1813/udp

VOLUME [ "/var/toughradius" ]
ENTRYPOINT ["/usr/bin/toughrad","start"]

