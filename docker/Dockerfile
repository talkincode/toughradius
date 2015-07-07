FROM centos:centos7
MAINTAINER jamiesun <jamiesun.net@gmail.com>


ADD ./radiusd.conf /etc/radiusd.conf
ADD ./supervisord.conf /etc/supervisord.conf
ADD ./toughrad /usr/bin/toughrad

RUN chmod +x /usr/bin/toughrad

RUN mkdir -p /var/toughradius/data

ADD ./privkey.pem /var/toughradius/privkey.pem
ADD ./cacert.pem /var/toughradius/cacert.pem

RUN yum update -y
RUN yum install -y libffi-devel openssl openssl-devel git gcc crontabs python-devel python-setuptools
RUN yum install -y  mysql-devel MySQL-python
RUN yum clean all

RUN easy_install pip
RUN pip install supervisor
RUN pip install Mako==0.9.0
RUN pip install Beaker==1.6.4
RUN pip install MarkupSafe==0.18
RUN pip install PyYAML==3.10
RUN pip install SQLAlchemy==0.9.8
RUN pip install Twisted==14.0.2
RUN pip install autobahn==0.9.3-3
RUN pip install bottle==0.12.7
RUN pip install six==1.8.0
RUN pip install tablib==0.10.0
RUN pip install zope.interface==4.1.1
RUN pip install pycrypto==2.6.1
RUN pip install pyOpenSSL>=0.14
RUN pip install service_identity

RUN git clone -b stable https://github.com/talkincode/ToughRADIUS.git /opt/toughradius
RUN ln -s /opt/toughradius/toughctl /usr/bin/toughctl && chmod +x /usr/bin/toughctl
RUN /opt/toughradius/toughctl --initdb

EXPOSE 1815 1816 1817 1819
EXPOSE 1812/udp
EXPOSE 1813/udp

ENTRYPOINT ["/usr/bin/toughrad","start"]

