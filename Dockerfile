# =============================================================================
# jamiesun/centos7-toughradius
#
# CentOS-7, MySQL5.6, Python2.7, ToughRADIUS
# 
# =============================================================================
FROM centos:centos7
MAINTAINER jamiesun <jamiesun.net@gmail.com>

VOLUME ["/var/toughradius"]

RUN mkdir -p /var/toughradius_run

ADD docker/my.cnf /etc/
ADD docker/supervisord.conf /var/toughradius_run/
ADD docker/radiusd.json /var/toughradius_run/
ADD docker/startup.sh /opt/
ADD docker/upgrade.sh /opt/
RUN chmod +x /opt/startup.sh   
RUN chmod +x /opt/upgrade.sh   

# install lib
RUN yum update -y
RUN yum install -y wget git gcc python-devel python-setuptools tcpdump

# install mysql
RUN rpm -ivh http://dev.mysql.com/get/mysql-community-release-el7-5.noarch.rpm
RUN yum install -y mysql-community-server mysql-community-devel 

# yum clean
RUN rm -rf /var/cache/yum/*
RUN yum clean all

RUN easy_install pip 
RUN easy_install supervisor

#install toughradius
RUN git clone https://github.com/talkincode/ToughRADIUS.git /opt/toughradius
RUN pip install -r /opt/toughradius/requirements.txt


EXPOSE 3306 1815 1816
EXPOSE 1812/udp
EXPOSE 1813/udp

ENTRYPOINT ["sh","/opt/startup.sh"]



