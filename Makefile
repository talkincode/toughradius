install:
	(\
	virtualenv venv --relocatable;\
	test -d /var/toughradius/data || mkdir -p /var/toughradius/data;\
	rm -f /etc/toughradius.conf && cp etc/toughradius.conf /etc/toughradius.conf;\
	test -f /etc/toughradius.json || cp etc/toughradius.json /etc/toughradius.json;\
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
	yum install -y czmq czmq-devel python-virtualenv supervisor;\
	yum install -y mysql-devel MySQL-python redis;\
	test -f /usr/local/bin/supervisord || ln -s `which supervisord` /usr/local/bin/supervisord;\
	test -f /usr/local/bin/supervisorctl || ln -s `which supervisorctl` /usr/local/bin/supervisorctl;\
	)

venv:
	(\
	test -d venv || virtualenv venv;\
	venv/bin/pip install -U pip;\
	venv/bin/pip install -U wheel;\
	venv/bin/pip install -U coverage;\
	venv/bin/pip install -U -r requirements.txt;\
	)

upgrade-libs:
	(\
	venv/bin/pip install -U --no-deps https://github.com/talkincode/toughlib/archive/master.zip;\
	venv/bin/pip install -U --no-deps https://github.com/talkincode/txradius/archive/master.zip;\
	)

upgrade-dev:
	git pull --rebase --stat origin release-dev

upgrade:
	git pull --rebase --stat origin release-stable

test:
	sh runtests.sh

initdb:
	python radiusctl initdb -f -c /etc/toughradius.json

inittest:
	python radiusctl inittest -c /etc/toughradius.json

clean:
	rm -fr venv

all:install-deps venv upgrade-libs install

.PHONY: all install install-deps upgrade-libs upgrade-dev upgrade test initdb inittest clean pypy pypy-initdb

