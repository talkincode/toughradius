FROM index.alauda.cn/toughstruct/tough-pypy:kiss
MAINTAINER jamiesun <jamiesun.net@gmail.com>

VOLUME [ "/var/toughradius" ]

ADD toughshell /usr/local/bin/toughshell
RUN chmod +x /usr/local/bin/toughshell
RUN /usr/local/bin/toughshell install

EXPOSE 1816
EXPOSE 1812/udp
EXPOSE 1813/udp

CMD ["/usr/local/bin/toughshell", "standalone"]

