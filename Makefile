install:
	(\
	virtualenv venv --relocatable;\
	test -d /var/toughradius/data || mkdir -p /var/toughradius/data;\
	rm -f /etc/toughradius.conf && cp etc/toughradius.conf /etc/toughradius.conf;\
	rm -f /etc/toughradius.json && cp etc/toughradius.json /etc/toughradius.json;\
	rm -f /etc/init.d/toughradius && cp etc/toughradius /etc/init.d/toughradius;\
	chmod +x /etc/init.d/toughradius && chkconfig toughradius on;\
	rm -f /usr/lib/systemd/system/toughradius.service && cp etc/toughradius.service /usr/lib/systemd/system/toughradius.service;\
	chmod 754 /usr/lib/systemd/system/toughradius.service && systemctl enable toughradius;\
	systemctl daemon-reload;\
	)

install-deps:
	(\
	yum install -y epel-release;\
	yum install -y wget zip python-devel libffi-devel openssl openssl-devel gcc git;\
	yum install -y czmq czmq-devel python-virtualenv;\
	yum install -y mysql-devel MySQL-python redis;\
	)

venv:
	(\
	test -d venv || virtualenv venv;\
	venv/bin/pip install -U pip;\
	venv/bin/pip install -U wheel;\
	venv/bin/pip install -U coverage;\
	test -d pymodules || mkdir pymodules;\
	venv/bin/pip download -d pymodules -r requirements.txt;\
	venv/bin/pip install -U --no-index --find-links=pymodules -r requirements.txt;\
	)

test:
	sh runtests.sh

initdb:
	venv/bin/python toughctl --initdb -f -c /etc/toughradius.json

inittest:
	venv/bin/python toughctl --inittest -c /etc/toughradius.json

clean:
	rm -fr pymodules  && rm -fr venv

all:install-deps venv install

.PHONY: all install install-deps initdb inittest