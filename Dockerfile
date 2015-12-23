FROM index.alauda.cn/toughstruct/tough-pypy:kiss
MAINTAINER jamiesun <jamiesun.net@gmail.com>

RUN git clone -b master https://github.com/talkincode/ToughRADIUS.git /opt/toughradius

RUN ln -s /opt/toughradius/toughradius.conf /etc/toughradius.conf

RUN chmod +x /opt/toughradius/toughctl

EXPOSE 1816
EXPOSE 18162
EXPOSE 18163

VOLUME [ "/var/toughradius" ]

CMD ["pypy", "/opt/toughradius/toughctl", "--admin"]

