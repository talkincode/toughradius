FROM talkincode/tough-pypy:kiss
MAINTAINER jamiesun <jamiesun.net@gmail.com>

RUN git clone -b master https://github.com/talkincode/toughengine.git /opt/toughengine

RUN chmod +x /opt/toughengine/toughctl

RUN ln -s /opt/toughengine/toughengine.conf /etc/toughengine.conf

EXPOSE 1815
EXPOSE 1812/udp
EXPOSE 1813/udp

VOLUME [ "/var/toughengine" ]

CMD ["pypy","/opt/toughengine/toughctl", "--standalone"]
