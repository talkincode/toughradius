FROM index.alauda.cn/toughstruct/tough-pypy:trv2
MAINTAINER jamiesun <jamiesun.net@gmail.com>

VOLUME [ "/var/toughradius" ]

RUN pypy -m pip install Click
RUN pypy -m pip install https://github.com/talkincode/toughlib/archive/master.zip --upgrade --no-deps
RUN pypy -m pip install https://github.com/talkincode/txradius/archive/master.zip --upgrade --no-deps
RUN git clone -b release-stable https://github.com/talkincode/ToughRADIUS.git /opt/toughradius && \
    ln -s /opt/toughradius/build/stable/toughradius.json /etc/toughradius.json && \
    ln -s /opt/toughradius/build/stable/supervisord.conf /etc/supervisord.conf && \
    chmod +x /opt/toughradius/radiusctl 

RUN wget http://qnstatic.toughcloud.net/toughkey_linux_x86_64 -O /usr/local/bin/toughkey && \
    chmod +x /usr/local/bin/toughkey

ADD entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 1816
EXPOSE 1812/udp
EXPOSE 1813/udp

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["/usr/local/bin/supervisord","-c","/etc/supervisord.conf"]


