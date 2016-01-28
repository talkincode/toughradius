FROM index.alauda.cn/toughstruct/tough-pypy:kiss
MAINTAINER jamiesun <jamiesun.net@gmail.com>

VOLUME [ "/var/toughradius" ]

ADD scripts/toughrun /usr/local/bin/toughrun
RUN chmod +x /usr/local/bin/toughrun
RUN /usr/local/bin/toughrun install

EXPOSE 1816
EXPOSE 1812/udp
EXPOSE 1813/udp

CMD ["/usr/local/bin/supervisord","-c","/etc/supervisord.conf"]

