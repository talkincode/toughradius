# =============================================================================
# jamiesun/centos7-toughradius
#
# CentOS-7, MySQL5.6, Python2.7, ToughRADIUS
# 
# =============================================================================
FROM centos:centos7
MAINTAINER jamiesun <jamiesun.net@gmail.com>

VOLUME ["/var/toughradius"]

RUN curl https://raw.githubusercontent.com/talkincode/ToughRADIUS/master/install/centos7-install.sh > centos7-install.sh
RUN sh centos7-install.sh depend
RUN sh centos7-install.sh mysql5
RUN sh centos7-install.sh radius
RUN sh centos7-install.sh setup
RUN /usr/bin/toughrad stop

EXPOSE 3306 1815 1816 1817
EXPOSE 1812/udp
EXPOSE 1813/udp

ENTRYPOINT ["/usr/bin/toughrad","start"]



