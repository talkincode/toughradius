FROM index.alauda.cn/toughstruct/tough-pypy:kiss
MAINTAINER jamiesun <jamiesun.net@gmail.com>

VOLUME [ "/var/toughradius" ]

RUN pypy -m pip install https://github.com/talkincode/toughlib/archive/master.zip --upgrade --no-deps

RUN git clone -b master https://github.com/talkincode/ToughRADIUS.git /opt/toughradius

RUN ln -s /opt/toughradius/toughradius.conf /etc/toughradius.conf

RUN chmod +x /opt/toughradius/toughctl

EXPOSE 1816
EXPOSE 18162
EXPOSE 18163

CMD ["pypy", "/opt/toughradius/toughctl", "--admin"]

