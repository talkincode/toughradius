FROM index.alauda.cn/toughstruct/tough-pypy
MAINTAINER jamiesun <jamiesun.net@gmail.com>

RUN git clone -b master https://github.com/talkincode/ToughRADIUS.git /opt/toughradius

RUN chmod +x /opt/toughradius/toughctl

RUN ln -s /opt/toughradius/etc/toughradius.conf /etc/toughradius.conf

RUN ln -s /opt/toughradius/etc/supervisord.conf /etc/supervisord.conf

RUN ln -s /opt/toughradius/bin/toughrad /usr/bin/toughrad && chmod +x /usr/bin/toughrad

EXPOSE 1816
EXPOSE 18162
EXPOSE 18163

VOLUME [ "/var/toughradius" ]
ENTRYPOINT ["/usr/bin/toughrad","start"]

