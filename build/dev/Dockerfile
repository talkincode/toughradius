FROM index.alauda.cn/toughstruct/tough-pypy:trv2
MAINTAINER jamiesun <jamiesun.net@gmail.com>

VOLUME [ "/var/toughradius" ]

ADD toughrun /usr/local/bin/toughrun
RUN chmod +x /usr/local/bin/toughrun
RUN /usr/local/bin/toughrun install

EXPOSE 1816

CMD ["pypy","/opt/toughradius/toughctl","--run"]

