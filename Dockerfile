FROM index.alauda.cn/toughstruct/tough-pypy:kiss
MAINTAINER jamiesun <jamiesun.net@gmail.com>

VOLUME [ "/var/toughradius" ]

RUN pypy -m pip install https://github.com/talkincode/toughlib/archive/master.zip --upgrade --no-deps
RUN pypy -m pip install https://github.com/talkincode/txradius/archive/master.zip --upgrade --no-deps

RUN git clone -b master https://github.com/talkincode/ToughRADIUS.git /opt/toughradius

RUN ln -s /opt/toughradius/toughradius.json /etc/toughradius.json

RUN chmod +x /opt/toughradius/toughctl

EXPOSE 1816
EXPOSE 1812/udp
EXPOSE 1813/udp

CMD ["pypy", "/opt/toughradius/toughctl", "--standalone"]

