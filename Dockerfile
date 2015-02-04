# =============================================================================
# jamiesun/centos7-toughradius
#
# CentOS-7, MySQL5.6, Python2.7, ToughRADIUS
# 
# =============================================================================
FROM centos:centos7
MAINTAINER jamiesun <jamiesun.net@gmail.com>

VOLUME ["/var/toughradius"]

RUN curl https://raw.githubusercontent.com/talkincode/ToughRADIUS/master/install/docker-install.sh > docker-install.sh
RUN sh docker-install.sh depend
RUN sh docker-install.sh mysql5
RUN sh docker-install.sh radius

EXPOSE 3306 1815 1816 1817
EXPOSE 1812/udp
EXPOSE 1813/udp

ENTRYPOINT ["/usr/bin/toughrad","docker"]



