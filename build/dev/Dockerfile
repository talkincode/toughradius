FROM index.alauda.cn/toughstruct/tough-pypy:trv2
MAINTAINER jamiesun <jamiesun.net@gmail.com>

VOLUME [ "/var/toughradius" ]

RUN pypy -m pip install https://github.com/talkincode/toughlib/archive/master.zip --upgrade --no-deps
RUN pypy -m pip install https://github.com/talkincode/txradius/archive/master.zip --upgrade --no-deps
RUN git clone -b master https://github.com/talkincode/ToughRADIUS.git /opt/toughradius && \
    ln -s /opt/toughradius/etc/toughradius.json /etc/toughradius.json && \
    ln -s /opt/toughradius/etc/supervisord.conf /etc/supervisord.conf && \
    chmod +x /opt/toughradius/toughctl 

ADD entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 1816
EXPOSE 1812/udp
EXPOSE 1813/udp

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["/usr/local/bin/supervisord","-c","/etc/supervisord.conf"]

